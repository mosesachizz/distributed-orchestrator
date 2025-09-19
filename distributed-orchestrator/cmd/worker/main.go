package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"distributed-orchestrator/pkg/util"
	"distributed-orchestrator/pkg/types"
	
	"go.etcd.io/etcd/client/v3"
)

type Worker struct {
	id      string
	address string
	store   store.Store
	client  *clientv3.Client
}

func NewWorker(etcdEndpoints []string) (*Worker, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	
	hostname, _ := os.Hostname()
	workerID := hostname + "-" + util.GenerateID()[:8]
	
	worker := &Worker{
		id:      workerID,
		address: getNodeIP(),
		client:  client,
	}
	
	// Initialize store
	worker.store, err = store.NewEtcdStore(etcdEndpoints, "orchestrator")
	if err != nil {
		return nil, err
	}
	
	return worker, nil
}

func (w *Worker) Run() {
	log.Printf("Worker %s starting on %s", w.id, w.address)
	
	// Register worker with master
	if err := w.register(); err != nil {
		log.Fatalf("Failed to register worker: %v", err)
	}
	
	// Start heartbeating
	go w.heartbeat()
	
	// Watch for assigned tasks
	go w.watchTasks()
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down worker...")
	w.deregister()
	w.client.Close()
}

func (w *Worker) register() error {
	node := &types.Node{
		ID:            w.id,
		Address:       w.address,
		State:         types.NodeStateReady,
		Capacity:      types.ResourceRequirements{CPU: 4000, Memory: 8192}, // 4 CPU, 8GB RAM
		Allocated:     types.ResourceRequirements{CPU: 0, Memory: 0},
		LastHeartbeat: time.Now().Unix(),
		CurrentTasks:  make(map[string]bool),
	}
	
	return w.store.StoreNode(context.Background(), node)
}

func (w *Worker) deregister() error {
	return w.store.DeleteNode(context.Background(), w.id)
}

func (w *Worker) heartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		node, err := w.store.GetNode(context.Background(), w.id)
		if err != nil {
			log.Printf("Failed to get node for heartbeat: %v", err)
			continue
		}
		
		node.LastHeartbeat = time.Now().Unix()
		if err := w.store.UpdateNode(context.Background(), node); err != nil {
			log.Printf("Failed to update heartbeat: %v", err)
		}
	}
}

func (w *Worker) watchTasks() {
	taskChan, err := w.store.WatchTasks(context.Background())
	if err != nil {
		log.Fatalf("Failed to watch tasks: %v", err)
	}
	
	for task := range taskChan {
		if task.NodeID == w.id && task.State == types.TaskStateAssigned {
			go w.executeTask(task)
		}
	}
}

func (w *Worker) executeTask(task *types.Task) {
	log.Printf("Executing task %s", task.ID)
	
	// Update task state to running
	task.State = types.TaskStateRunning
	task.StartTime = time.Now().Unix()
	if err := w.store.UpdateTask(context.Background(), task); err != nil {
		log.Printf("Failed to update task state: %v", err)
		return
	}
	
	// Execute the task using Docker
	output, exitCode, err := util.ExecuteDockerTask(task)
	
	// Update task state based on execution result
	task.EndTime = time.Now().Unix()
	task.ExitCode = exitCode
	
	if err != nil {
		log.Printf("Task %s failed: %v", task.ID, err)
		task.State = types.TaskStateFailed
	} else if exitCode == 0 {
		task.State = types.TaskStateCompleted
		log.Printf("Task %s completed successfully", task.ID)
	} else {
		task.State = types.TaskStateFailed
		log.Printf("Task %s failed with exit code %d", task.ID, exitCode)
	}
	
	// Store task result
	if err := w.store.UpdateTask(context.Background(), task); err != nil {
		log.Printf("Failed to update task result: %v", err)
	}
	
	// Update node resource allocation
	if node, err := w.store.GetNode(context.Background(), w.id); err == nil {
		node.ReleaseResources(task)
		if err := w.store.UpdateNode(context.Background(), node); err != nil {
			log.Printf("Failed to update node resources: %v", err)
		}
	}
}

func getNodeIP() string {
	// Simplified implementation - in production, you'd want a more robust method
	if ip := os.Getenv("NODE_IP"); ip != "" {
		return ip
	}
	return "localhost"
}

func main() {
	etcdEndpoints := []string{"localhost:2379"}
	if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
		etcdEndpoints = []string{endpoints}
	}
	
	worker, err := NewWorker(etcdEndpoints)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	
	worker.Run()
}