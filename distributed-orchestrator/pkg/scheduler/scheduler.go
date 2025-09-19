package scheduler

import (
	"context"
	"log"
	"time"
	
	"distributed-orchestrator/pkg/store"
	"distributed-orchestrator/pkg/types"
	"distributed-orchestrator/pkg/util"
)

type Scheduler struct {
	store        store.Store
	algorithm    SchedulingAlgorithm
	eventChannel chan *types.Event
}

type SchedulingAlgorithm interface {
	SelectNode(task *types.Task, nodes []*types.Node) *types.Node
}

func NewScheduler(store store.Store, algorithm SchedulingAlgorithm) *Scheduler {
	return &Scheduler{
		store:        store,
		algorithm:    algorithm,
		eventChannel: make(chan *types.Event, 100),
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	log.Println("Scheduler started")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopping")
			return
		case <-ticker.C:
			s.reconcile(ctx)
		}
	}
}

func (s *Scheduler) reconcile(ctx context.Context) {
	// Get all pending tasks
	tasks, err := s.store.GetAllTasks(ctx)
	if err != nil {
		log.Printf("Failed to get tasks: %v", err)
		return
	}
	
	// Get all available nodes
	nodes, err := s.store.GetAllNodes(ctx)
	if err != nil {
		log.Printf("Failed to get nodes: %v", err)
		return
	}
	
	// Filter ready nodes
	var readyNodes []*types.Node
	for _, node := range nodes {
		if node.State == types.NodeStateReady && !node.Draining {
			readyNodes = append(readyNodes, node)
		}
	}
	
	// Schedule pending tasks
	for _, task := range tasks {
		if task.State == types.TaskStatePending {
			selectedNode := s.algorithm.SelectNode(task, readyNodes)
			if selectedNode != nil {
				if err := s.assignTaskToNode(ctx, task, selectedNode); err != nil {
					log.Printf("Failed to assign task %s to node %s: %v", task.ID, selectedNode.ID, err)
				} else {
					log.Printf("Assigned task %s to node %s", task.ID, selectedNode.ID)
				}
			} else {
				log.Printf("No suitable node found for task %s", task.ID)
			}
		}
	}
	
	// Handle node failures and task timeouts
	s.checkNodeHealth(ctx, nodes)
	s.checkTaskTimeouts(ctx, tasks)
}

func (s *Scheduler) assignTaskToNode(ctx context.Context, task *types.Task, node *types.Node) error {
	// Update task state
	task.State = types.TaskStateAssigned
	task.NodeID = node.ID
	
	// Update node resource allocation
	node.AllocateResources(task)
	
	// Persist changes
	if err := s.store.UpdateTask(ctx, task); err != nil {
		return err
	}
	if err := s.store.UpdateNode(ctx, node); err != nil {
		return err
	}
	
	// Emit event
	s.emitEvent(&types.Event{
		Type: "TaskAssigned",
		Object: map[string]interface{}{
			"task": task,
			"node": node,
		},
		Timestamp: time.Now().Unix(),
	})
	
	return nil
}

func (s *Scheduler) checkNodeHealth(ctx context.Context, nodes []*types.Node) {
	now := time.Now().Unix()
	
	for _, node := range nodes {
		// If node hasn't heartbeated in 30 seconds, mark it as not ready
		if now-node.LastHeartbeat > 30 {
			if node.State == types.NodeStateReady {
				log.Printf("Node %s is not responding, marking as not ready", node.ID)
				node.State = types.NodeStateNotReady
				if err := s.store.UpdateNode(ctx, node); err != nil {
					log.Printf("Failed to update node %s: %v", node.ID, err)
				}
				
				// Reschedule tasks from failed node
				s.rescheduleTasksFromNode(ctx, node)
			}
		} else if node.State == types.NodeStateNotReady {
			// Node recovered
			log.Printf("Node %s recovered, marking as ready", node.ID)
			node.State = types.NodeStateReady
			if err := s.store.UpdateNode(ctx, node); err != nil {
				log.Printf("Failed to update node %s: %v", node.ID, err)
			}
		}
	}
}

func (s *Scheduler) checkTaskTimeouts(ctx context.Context, tasks []*types.Task) {
	now := time.Now().Unix()
	
	for _, task := range tasks {
		// If task has been running for more than 1 hour, consider it timed out
		if task.State == types.TaskStateRunning && now-task.StartTime > 3600 {
			log.Printf("Task %s timed out", task.ID)
			task.State = types.TaskStateFailed
			task.ExitCode = -1
			task.EndTime = now
			
			if err := s.store.UpdateTask(ctx, task); err != nil {
				log.Printf("Failed to update task %s: %v", task.ID, err)
			}
			
			// Release resources on the node
			if node, err := s.store.GetNode(ctx, task.NodeID); err == nil {
				node.ReleaseResources(task)
				if err := s.store.UpdateNode(ctx, node); err != nil {
					log.Printf("Failed to update node %s: %v", node.ID, err)
				}
			}
		}
	}
}

func (s *Scheduler) rescheduleTasksFromNode(ctx context.Context, node *types.Node) {
	tasks, err := s.store.GetAllTasks(ctx)
	if err != nil {
		log.Printf("Failed to get tasks for rescheduling: %v", err)
		return
	}
	
	for _, task := range tasks {
		if task.NodeID == node.ID && 
		   (task.State == types.TaskStateAssigned || task.State == types.TaskStateRunning) {
			// Reset task to pending state for rescheduling
			task.State = types.TaskStatePending
			task.NodeID = ""
			
			if err := s.store.UpdateTask(ctx, task); err != nil {
				log.Printf("Failed to reset task %s: %v", task.ID, err)
			}
			
			// Release resources from the failed node
			node.ReleaseResources(task)
		}
	}
	
	// Update the node after releasing all resources
	if err := s.store.UpdateNode(ctx, node); err != nil {
		log.Printf("Failed to update node %s: %v", node.ID, err)
	}
}

func (s *Scheduler) emitEvent(event *types.Event) {
	event.ID = util.GenerateID()
	select {
	case s.eventChannel <- event:
	default:
		log.Println("Event channel full, dropping event")
	}
}

func (s *Scheduler) GetEventChannel() <-chan *types.Event {
	return s.eventChannel
}