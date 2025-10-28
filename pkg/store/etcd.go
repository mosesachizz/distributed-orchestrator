package store

import (
	"context"
	"encoding/json"
	"log"
	"time"
	
	"distributed-orchestrator/pkg/types"
	"go.etcd.io/etcd/client/v3"
)

type EtcdStore struct {
	client *clientv3.Client
	prefix string
}

func NewEtcdStore(endpoints []string, prefix string) (*EtcdStore, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	
	return &EtcdStore{
		client: client,
		prefix: prefix,
	}, nil
}

func (s *EtcdStore) key(entityType, id string) string {
	return s.prefix + "/" + entityType + "/" + id
}

func (s *EtcdStore) StoreTask(ctx context.Context, task *types.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	
	_, err = s.client.Put(ctx, s.key("tasks", task.ID), string(data))
	return err
}

func (s *EtcdStore) GetTask(ctx context.Context, id string) (*types.Task, error) {
	resp, err := s.client.Get(ctx, s.key("tasks", id))
	if err != nil {
		return nil, err
	}
	
	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}
	
	var task types.Task
	if err := json.Unmarshal(resp.Kvs[0].Value, &task); err != nil {
		return nil, err
	}
	
	return &task, nil
}

func (s *EtcdStore) GetAllTasks(ctx context.Context) ([]*types.Task, error) {
	resp, err := s.client.Get(ctx, s.key("tasks", ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	
	var tasks []*types.Task
	for _, kv := range resp.Kvs {
		var task types.Task
		if err := json.Unmarshal(kv.Value, &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			continue
		}
		tasks = append(tasks, &task)
	}
	
	return tasks, nil
}

func (s *EtcdStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return s.StoreTask(ctx, task)
}

func (s *EtcdStore) DeleteTask(ctx context.Context, id string) error {
	_, err := s.client.Delete(ctx, s.key("tasks", id))
	return err
}

func (s *EtcdStore) StoreNode(ctx context.Context, node *types.Node) error {
	data, err := json.Marshal(node)
	if err != nil {
		return err
	}
	
	_, err = s.client.Put(ctx, s.key("nodes", node.ID), string(data))
	return err
}

func (s *EtcdStore) GetNode(ctx context.Context, id string) (*types.Node, error) {
	resp, err := s.client.Get(ctx, s.key("nodes", id))
	if err != nil {
		return nil, err
	}
	
	if len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}
	
	var node types.Node
	if err := json.Unmarshal(resp.Kvs[0].Value, &node); err != nil {
		return nil, err
	}
	
	return &node, nil
}

func (s *EtcdStore) GetAllNodes(ctx context.Context) ([]*types.Node, error) {
	resp, err := s.client.Get(ctx, s.key("nodes", ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	
	var nodes []*types.Node
	for _, kv := range resp.Kvs {
		var node types.Node
		if err := json.Unmarshal(kv.Value, &node); err != nil {
			log.Printf("Failed to unmarshal node: %v", err)
			continue
		}
		nodes = append(nodes, &node)
	}
	
	return nodes, nil
}

func (s *EtcdStore) UpdateNode(ctx context.Context, node *types.Node) error {
	return s.StoreNode(ctx, node)
}

func (s *EtcdStore) DeleteNode(ctx context.Context, id string) error {
	_, err := s.client.Delete(ctx, s.key("nodes", id))
	return err
}

func (s *EtcdStore) WatchTasks(ctx context.Context) (<-chan *types.Task, error) {
	watchChan := s.client.Watch(ctx, s.key("tasks", ""), clientv3.WithPrefix())
	taskChan := make(chan *types.Task, 10)
	
	go func() {
		defer close(taskChan)
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				if event.Type == clientv3.EventTypePut {
					var task types.Task
					if err := json.Unmarshal(event.Kv.Value, &task); err == nil {
						taskChan <- &task
					}
				}
			}
		}
	}()
	
	return taskChan, nil
}

func (s *EtcdStore) Close() error {
	return s.client.Close()
}