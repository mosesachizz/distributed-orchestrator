package rpc

import (
	"context"
	"log"
	"time"
	
	pb "distributed-orchestrator/pkg/rpc/proto"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WorkerClient struct {
	client pb.OrchestratorClient
	conn   *grpc.ClientConn
	nodeID string
}

func NewWorkerClient(masterAddr, nodeID string) (*WorkerClient, error) {
	conn, err := grpc.Dial(masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	
	client := pb.NewOrchestratorClient(conn)
	return &WorkerClient{
		client: client,
		conn:   conn,
		nodeID: nodeID,
	}, nil
}

func (c *WorkerClient) Close() error {
	return c.conn.Close()
}

func (c *WorkerClient) Register(capacity *pb.Resource, labels map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := c.client.RegisterWorker(ctx, &pb.RegisterWorkerRequest{
		NodeId:   c.nodeID,
		Address:  getNodeAddress(),
		Capacity: capacity,
		Labels:   labels,
	})
	return err
}

func (c *WorkerClient) SendHeartbeat(allocated *pb.Resource, runningTasks []string) ([]*pb.TaskAssignment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	resp, err := c.client.Heartbeat(ctx, &pb.HeartbeatRequest{
		NodeId:       c.nodeID,
		Allocated:    allocated,
		RunningTasks: runningTasks,
	})
	if err != nil {
		return nil, err
	}
	
	return resp.Assignments, nil
}

func (c *WorkerClient) SendTaskUpdate(taskID string, state pb.TaskState, exitCode int32, output string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := c.client.TaskUpdate(ctx, &pb.TaskUpdateRequest{
		TaskId:   taskID,
		NodeId:   c.nodeID,
		State:    state,
		ExitCode: exitCode,
		Output:   output,
	})
	return err
}

func (c *WorkerClient) GetNodeStats() (*pb.GetNodeStatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return c.client.GetNodeStats(ctx, &pb.GetNodeStatsRequest{NodeId: c.nodeID})
}

func getNodeAddress() string {
	// Implementation to get node's IP address
	return "localhost" // Simplified - use actual IP detection
}