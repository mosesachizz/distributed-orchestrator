package scheduler

import (
	"math/rand"
	"time"
	
	"distributed-orchestrator/pkg/types"
)

// RoundRobinAlgorithm implements round-robin scheduling
type RoundRobinAlgorithm struct {
	lastNodeIndex int
}

func NewRoundRobinAlgorithm() *RoundRobinAlgorithm {
	rand.Seed(time.Now().UnixNano())
	return &RoundRobinAlgorithm{lastNodeIndex: -1}
}

func (a *RoundRobinAlgorithm) SelectNode(task *types.Task, nodes []*types.Node) *types.Node {
	if len(nodes) == 0 {
		return nil
	}
	
	// Filter nodes that have enough resources
	var suitableNodes []*types.Node
	for _, node := range nodes {
		if node.HasResourcesFor(task) {
			suitableNodes = append(suitableNodes, node)
		}
	}
	
	if len(suitableNodes) == 0 {
		return nil
	}
	
	// Round-robin selection
	a.lastNodeIndex = (a.lastNodeIndex + 1) % len(suitableNodes)
	return suitableNodes[a.lastNodeIndex]
}

// BinPackingAlgorithm implements bin packing for better resource utilization
type BinPackingAlgorithm struct{}

func NewBinPackingAlgorithm() *BinPackingAlgorithm {
	return &BinPackingAlgorithm{}
}

func (a *BinPackingAlgorithm) SelectNode(task *types.Task, nodes []*types.Node) *types.Node {
	var bestNode *types.Node
	bestScore := -1.0
	
	for _, node := range nodes {
		if !node.HasResourcesFor(task) {
			continue
		}
		
		// Calculate resource utilization score (weighted average of CPU and memory)
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

// SpreadAlgorithm implements spread scheduling for better fault tolerance
type SpreadAlgorithm struct{}

func NewSpreadAlgorithm() *SpreadAlgorithm {
	return &SpreadAlgorithm{}
}

func (a *SpreadAlgorithm) SelectNode(task *types.Task, nodes []*types.Node) *types.Node {
	var bestNode *types.Node
	bestScore := -1.0
	
	for _, node := range nodes {
		if !node.HasResourcesFor(task) {
			continue
		}
		
		// Calculate score based on inverse of current task count
		// This spreads tasks across nodes
		score := 1.0 / float64(len(node.CurrentTasks)+1)
		
		if bestNode == nil || score > bestScore {
			bestNode = node
			bestScore = score
		}
	}
	
	return bestNode
}