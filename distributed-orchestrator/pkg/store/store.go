package store

import (
	"context"
	"distributed-orchestrator/pkg/types"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	// Task operations
	StoreTask(ctx context.Context, task *types.Task) error
	GetTask(ctx context.Context, id string) (*types.Task, error)
	GetAllTasks(ctx context.Context) ([]*types.Task, error)
	UpdateTask(ctx context.Context, task *types.Task) error
	DeleteTask(ctx context.Context, id string) error
	WatchTasks(ctx context.Context) (<-chan *types.Task, error)

	// Node operations
	StoreNode(ctx context.Context, node *types.Node) error
	GetNode(ctx context.Context, id string) (*types.Node, error)
	GetAllNodes(ctx context.Context) ([]*types.Node, error)
	UpdateNode(ctx context.Context, node *types.Node) error
	DeleteNode(ctx context.Context, id string) error
	WatchNodes(ctx context.Context) (<-chan *types.Node, error)

	// Event operations
	StoreEvent(ctx context.Context, event *types.Event) error
	GetEvents(ctx context.Context, limit int) ([]*types.Event, error)

	// Close the store
	Close() error
}