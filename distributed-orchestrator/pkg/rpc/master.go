package rpc

import (
	"context"
	"log"
	
	"distributed-orchestrator/pkg/store"
	"distributed-orchestrator/pkg/types"
	pb "distributed-orchestrator/pkg/rpc/proto"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterServer struct {
	pb.UnimplementedOrchestratorServer
	store store.Store
}

func NewMasterServer(store store.Store) *MasterServer {
	return &MasterServer{store: store}
}

func (s *MasterServer) RegisterWorker(ctx context.Context, req *pb.RegisterWorkerRequest) (*pb.RegisterWorkerResponse, error) {
	node := &types.Node{
		ID:            req.NodeId,
		Address:       req.Address,
		State:         types.NodeStateReady,
		Capacity:      types.ResourceRequirements{CPU: req.Capacity.Cpu, Memory: req.Capacity.Memory},
		Allocated:     types.ResourceRequirements{CPU: 0, Memory: 0},
		LastHeartbeat: types.NowUnix(),
		CurrentTasks:  make(map[string]bool),
		Labels:        req.Labels,
	}
	
	if err := s.store.StoreNode(ctx, node); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register worker: %v", err)
	}
	
	log.Printf("Worker registered: %s (%s)", req.NodeId, req.Address)
	return &pb.RegisterWorkerResponse{Success: true, Message: "Worker registered successfully"}, nil
}

func (s *MasterServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	node, err := s.store.GetNode(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "worker not found: %s", req.NodeId)
	}
	
	// Update node state
	node.LastHeartbeat = types.NowUnix()
	node.Allocated = types.ResourceRequirements{CPU: req.Allocated.Cpu, Memory: req.Allocated.Memory}
	
	// Update current tasks
	node.CurrentTasks = make(map[string]bool)
	for _, taskID := range req.RunningTasks {
		node.CurrentTasks[taskID] = true
	}
	
	if err := s.store.UpdateNode(ctx, node); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update worker: %v", err)
	}
	
	// Get any new task assignments
	tasks, err := s.store.GetAllTasks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get tasks: %v", err)
	}
	
	var assignments []*pb.TaskAssignment
	for _, task := range tasks {
		if task.State == types.TaskStatePending && node.HasResourcesFor(task) {
			assignments = append(assignments, &pb.TaskAssignment{
				TaskId: task.ID,
				Task: &pb.TaskSpec{
					Id:         task.ID,
					Name:       task.Name,
					Image:      task.Image,
					Command:    task.Command,
					Env:        task.Env,
					Resources: &pb.Resource{
						Cpu:    task.Resources.CPU,
						Memory: task.Resources.Memory,
					},
				},
			})
			
			// Update task state
			task.State = types.TaskStateAssigned
			task.NodeID = node.ID
			if err := s.store.UpdateTask(ctx, task); err != nil {
				log.Printf("Failed to update task %s: %v", task.ID, err)
			}
		}
	}
	
	return &pb.HeartbeatResponse{
		Success:    true,
		Assignments: assignments,
	}, nil
}

func (s *MasterServer) TaskUpdate(ctx context.Context, req *pb.TaskUpdateRequest) (*pb.TaskUpdateResponse, error) {
	task, err := s.store.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %s", req.TaskId)
	}
	
	// Update task state
	switch req.State {
	case pb.TaskState_RUNNING:
		task.State = types.TaskStateRunning
		task.StartTime = types.NowUnix()
	case pb.TaskState_COMPLETED:
		task.State = types.TaskStateCompleted
		task.EndTime = types.NowUnix()
		task.ExitCode = int(req.ExitCode)
	case pb.TaskState_FAILED:
		task.State = types.TaskStateFailed
		task.EndTime = types.NowUnix()
		task.ExitCode = int(req.ExitCode)
	}
	
	if err := s.store.UpdateTask(ctx, task); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}
	
	log.Printf("Task %s updated to state %s", req.TaskId, req.State)
	return &pb.TaskUpdateResponse{Success: true}, nil
}

func (s *MasterServer) GetNodeStats(ctx context.Context, req *pb.GetNodeStatsRequest) (*pb.GetNodeStatsResponse, error) {
	node, err := s.store.GetNode(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "node not found: %s", req.NodeId)
	}
	
	return &pb.GetNodeStatsResponse{
		Capacity: &pb.Resource{
			Cpu:    node.Capacity.CPU,
			Memory: node.Capacity.Memory,
		},
		Allocated: &pb.Resource{
			Cpu:    node.Allocated.CPU,
			Memory: node.Allocated.Memory,
		},
		TotalTasks:   len(node.CurrentTasks),
		RunningTasks: len(node.CurrentTasks), // Simplified - in reality you'd check task states
	}, nil
}

func RegisterMasterServer(grpcServer *grpc.Server, store store.Store) {
	pb.RegisterOrchestratorServer(grpcServer, NewMasterServer(store))
}