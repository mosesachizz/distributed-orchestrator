package store

import (
	"context"
	"sync"
	"time"
	
	"distributed-orchestrator/pkg/types"
)

type MemoryStore struct {
	mu     sync.RWMutex
	tasks  map[string]*types.Task
	nodes  map[string]*types.Node
	events []*types.Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks:  make(map[string]*types.Task),
		nodes:  make(map[string]*types.Node),
		events: make([]*types.Event, 0),
	}
}

func (s *MemoryStore) StoreTask(ctx context.Context, task *types.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStore) GetTask(ctx context.Context, id string) (*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrNotFound
	}
	return task, nil
}

func (s *MemoryStore) GetAllTasks(ctx context.Context) ([]*types.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks := make([]*types.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *MemoryStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return s.StoreTask(ctx, task)
}

func (s *MemoryStore) DeleteTask(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.tasks, id)
	return nil
}

func (s *MemoryStore) StoreNode(ctx context.Context, node *types.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.nodes[node.ID] = node
	return nil
}

func (s *MemoryStore) GetNode(ctx context.Context, id string) (*types.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	node, exists := s.nodes[id]
	if !exists {
		return nil, ErrNotFound
	}
	return node, nil
}

func (s *MemoryStore) GetAllNodes(ctx context.Context) ([]*types.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	nodes := make([]*types.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *MemoryStore) UpdateNode(ctx context.Context, node *types.Node) error {
	return s.StoreNode(ctx, node)
}

func (s *MemoryStore) DeleteNode(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.nodes, id)
	return nil
}

func (s *MemoryStore) StoreEvent(ctx context.Context, event *types.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.events = append(s.events, event)
	if len(s.events) > 1000 {
		s.events = s.events[len(s.events)-1000:]
	}
	return nil
}

func (s *MemoryStore) GetEvents(ctx context.Context, limit int) ([]*types.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if limit > len(s.events) {
		limit = len(s.events)
	}
	return s.events[len(s.events)-limit:], nil
}

func (s *MemoryStore) WatchTasks(ctx context.Context) (<-chan *types.Task, error) {
	// Memory store doesn't support watching
	taskChan := make(chan *types.Task, 10)
	close(taskChan)
	return taskChan, nil
}

func (s *MemoryStore) WatchNodes(ctx context.Context) (<-chan *types.Node, error) {
	// Memory store doesn't support watching
	nodeChan := make(chan *types.Node, 10)
	close(nodeChan)
	return nodeChan, nil
}

func (s *MemoryStore) Close() error {
	return nil
}