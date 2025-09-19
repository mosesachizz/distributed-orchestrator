# Distributed Task Orchestrator

A Kubernetes-inspired distributed task orchestrator that schedules and manages containerized tasks across a cluster of worker nodes.

## Problem Statement

Managing containerized workloads across multiple machines requires sophisticated scheduling, resource management, and fault tolerance. Manual management becomes impractical at scale.

## Solution

This orchestrator provides:
- Automated scheduling of containerized tasks
- Resource allocation and management
- Health monitoring and self-healing
- Horizontal scaling of worker nodes
- REST API for task management

## Tech Stack

- **Language**: Go (for performance and concurrency)
- **Storage**: etcd (for distributed consensus and state storage)
- **Container Runtime**: Docker
- **Communication**: gRPC (for internal communication), HTTP REST API
- **Scheduling Algorithms**: Multiple algorithms (Round Robin, Bin Packing, Spread)
- **Deployment**: Docker, Docker Compose

## Architecture Decisions

1. **Master-Worker Architecture**: Centralized scheduling with distributed execution
2. **etcd for State Storage**: Provides strong consistency for cluster state
3. **Multiple Scheduling Algorithms**: Different strategies for different workload types
4. **Health Checking**: Automated node health monitoring and task rescheduling
5. **Resource Management**: Fine-grained CPU and memory allocation

## Key Features

### Scheduling Algorithm (Bin Packing)

```go
func (a *BinPackingAlgorithm) SelectNode(task *types.Task, nodes []*types.Node) *types.Node {
	var bestNode *types.Node
	bestScore := -1.0
	
	for _, node := range nodes {
		if !node.HasResourcesFor(task) {
			continue
		}
		
		// Calculate resource utilization score
		cpuUtilization := float64(node.Allocated.CPU+task.Resources.CPU) / float64(node.Capacity.CPU)
		memoryUtilization := float64(node.Allocated.Memory+task.Resources.Memory) / float64(node.Capacity.Memory)
		score := 0.7*cpuUtilization + 0.3*memoryUtilization
		
		if bestNode == nil || score > bestScore {
			bestNode = node
			bestScore = score
		}
	}
	
	return bestNode
}

## Setup Instructions
Prerequisites
• Docker and Docker Compose
• Go 1.18+ (for development)

Running with Docker Compose
1. **Clone the repository**:
git clone https://github.com/your-username/distributed-orchestrator.git
cd distributed-orchestrator

2. **Start the cluster**:
docker-compose up -d

3. **Access the API**:
• Master API: http://localhost:8080
• etcd: http://localhost:2379

Building from Source
1. **Build the master**:
go build -o bin/master ./cmd/master

2. **Build the worker**:
go build -o bin/worker ./cmd/worker

3. **Start etcd**:
docker run -d -p 2379:2379 --name etcd quay.io/coreos/etcd:v3.5.0

4. **Start the master**:
./bin/master

5. **Start workers (in separate terminals)**:
./bin/worker

## API Usage
**Create a task**:
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-task",
    "image": "alpine",
    "command": ["echo", "hello world"],
    "resources": {
      "cpu": 500,
      "memory": 256
    }
  }'

**List tasks**:
curl http://localhost:8080/api/v1/tasks
List nodes:
curl http://localhost:8080/api/v1/nodes

## Monitoring
The system provides event streaming at /api/v1/events for real-time monitoring of cluster activities.

## Scaling
**To add more workers**:
docker-compose up -d --scale worker=3

# License
This project is licensed under the MIT License - see the LICENSE file for details.