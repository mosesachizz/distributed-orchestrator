package types

type TaskState string
type NodeState string

const (
	TaskStatePending   TaskState = "PENDING"
	TaskStateAssigned  TaskState = "ASSIGNED"
	TaskStateRunning   TaskState = "RUNNING"
	TaskStateCompleted TaskState = "COMPLETED"
	TaskStateFailed    TaskState = "FAILED"
	
	NodeStateReady    NodeState = "READY"
	NodeStateNotReady NodeState = "NOT_READY"
	NodeStateDraining NodeState = "DRAINING"
)

type ResourceRequirements struct {
	CPU    int64 `json:"cpu"`    // in millicores
	Memory int64 `json:"memory"` // in megabytes
}

type Task struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Image        string               `json:"image"`
	Command      []string             `json:"command,omitempty"`
	Env          map[string]string    `json:"env,omitempty"`
	Resources    ResourceRequirements `json:"resources"`
	State        TaskState            `json:"state"`
	NodeID       string               `json:"node_id,omitempty"`
	CreationTime int64                `json:"creation_time"`
	StartTime    int64                `json:"start_time,omitempty"`
	EndTime      int64                `json:"end_time,omitempty"`
	ExitCode     int                  `json:"exit_code,omitempty"`
}

type Node struct {
	ID              string               `json:"id"`
	Address         string               `json:"address"`
	State           NodeState            `json:"state"`
	Capacity        ResourceRequirements `json:"capacity"`
	Allocated       ResourceRequirements `json:"allocated"`
	Labels          map[string]string    `json:"labels,omitempty"`
	LastHeartbeat   int64                `json:"last_heartbeat"`
	CurrentTasks    map[string]bool      `json:"current_tasks"` // taskID -> true
	Draining        bool                 `json:"draining"`
}

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Object    map[string]interface{} `json:"object"`
	Timestamp int64                  `json:"timestamp"`
}

func (n *Node) HasResourcesFor(task *Task) bool {
	availableCPU := n.Capacity.CPU - n.Allocated.CPU
	availableMemory := n.Capacity.Memory - n.Allocated.Memory
	
	return availableCPU >= task.Resources.CPU && 
	       availableMemory >= task.Resources.Memory &&
	       n.State == NodeStateReady &&
	       !n.Draining
}

func (n *Node) AllocateResources(task *Task) {
	n.Allocated.CPU += task.Resources.CPU
	n.Allocated.Memory += task.Resources.Memory
	n.CurrentTasks[task.ID] = true
}

func (n *Node) ReleaseResources(task *Task) {
	n.Allocated.CPU -= task.Resources.CPU
	n.Allocated.Memory -= task.Resources.Memory
	delete(n.CurrentTasks, task.ID)
}