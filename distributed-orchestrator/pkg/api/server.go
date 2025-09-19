package api

import (
	"encoding/json"
	"net/http"
	"time"
	
	"distributed-orchestrator/pkg/scheduler"
	"distributed-orchestrator/pkg/store"
	"distributed-orchestrator/pkg/types"
	
	"github.com/gorilla/mux"
)

type Server struct {
	store     store.Store
	scheduler *scheduler.Scheduler
}

func NewServer(store store.Store, scheduler *scheduler.Scheduler) *Server {
	return &Server{
		store:     store,
		scheduler: scheduler,
	}
}

func (s *Server) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/tasks", s.createTask).Methods("POST")
	router.HandleFunc("/api/v1/tasks", s.listTasks).Methods("GET")
	router.HandleFunc("/api/v1/tasks/{id}", s.getTask).Methods("GET")
	router.HandleFunc("/api/v1/tasks/{id}", s.deleteTask).Methods("DELETE")
	
	router.HandleFunc("/api/v1/nodes", s.listNodes).Methods("GET")
	router.HandleFunc("/api/v1/nodes/{id}", s.getNode).Methods("GET")
	router.HandleFunc("/api/v1/nodes/{id}/drain", s.drainNode).Methods("POST")
	
	router.HandleFunc("/api/v1/events", s.watchEvents).Methods("GET")
	
	router.HandleFunc("/health", s.healthCheck).Methods("GET")
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var task types.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}
	
	// Validate task
	if task.Image == "" {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}
	
	// Set task metadata
	task.ID = util.GenerateID()
	task.State = types.TaskStatePending
	task.CreationTime = time.Now().Unix()
	
	// Set default resources if not specified
	if task.Resources.CPU == 0 {
		task.Resources.CPU = 1000 // 1 CPU core
	}
	if task.Resources.Memory == 0 {
		task.Resources.Memory = 512 // 512MB
	}
	
	// Store task
	if err := s.store.StoreTask(r.Context(), &task); err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.store.GetAllTasks(r.Context())
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	
	task, err := s.store.GetTask(r.Context(), taskID)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	
	// Get task first to check if it's running
	task, err := s.store.GetTask(r.Context(), taskID)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}
	
	// If task is running on a node, we need to stop it and release resources
	if task.NodeID != "" && (task.State == types.TaskStateAssigned || task.State == types.TaskStateRunning) {
		node, err := s.store.GetNode(r.Context(), task.NodeID)
		if err == nil {
			node.ReleaseResources(task)
			if err := s.store.UpdateNode(r.Context(), node); err != nil {
				log.Printf("Failed to update node resources: %v", err)
			}
		}
	}
	
	if err := s.store.DeleteTask(r.Context(), taskID); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.store.GetAllNodes(r.Context())
	if err != nil {
		http.Error(w, "Failed to get nodes", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

func (s *Server) getNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeID := vars["id"]
	
	node, err := s.store.GetNode(r.Context(), nodeID)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "Node not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get node", http.StatusInternalServerError)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

func (s *Server) drainNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeID := vars["id"]
	
	node, err := s.store.GetNode(r.Context(), nodeID)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "Node not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get node", http.StatusInternalServerError)
		}
		return
	}
	
	// Mark node as draining
	node.Draining = true
	if err := s.store.UpdateNode(r.Context(), node); err != nil {
		http.Error(w, "Failed to update node", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "node draining started"})
}

func (s *Server) watchEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	
	eventChan := s.scheduler.GetEventChannel()
	
	for {
		select {
		case event := <-eventChan:
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Failed to marshal event: %v", err)
				continue
			}
			
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}