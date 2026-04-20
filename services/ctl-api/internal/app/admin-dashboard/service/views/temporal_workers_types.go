package views

import (
	"time"
)

// NamespaceWorkerInfo holds worker/poller information for a single Temporal namespace.
type NamespaceWorkerInfo struct {
	Namespace string
	TaskQueue string
	Error     string // non-empty if we failed to query this namespace

	WorkflowPollers []PollerDetail
	ActivityPollers []PollerDetail
	WorkflowStats   *TaskQueueStatsInfo
	ActivityStats   *TaskQueueStatsInfo
}

// PollerDetail represents a single worker/poller connected to a task queue.
type PollerDetail struct {
	Identity       string
	LastAccessTime time.Time
	RatePerSecond  float64
}

// TaskQueueStatsInfo holds stats about a task queue.
type TaskQueueStatsInfo struct {
	ApproximateBacklogCount int64
	ApproximateBacklogAge   time.Duration
	TasksAddRate            float32
	TasksDispatchRate       float32
}

// TotalPollerCount returns the total number of pollers across workflow and activity queues.
func (n *NamespaceWorkerInfo) TotalPollerCount() int {
	return len(n.WorkflowPollers) + len(n.ActivityPollers)
}

// IsHealthy returns true if there are active pollers and no error.
func (n *NamespaceWorkerInfo) IsHealthy() bool {
	return n.Error == "" && (len(n.WorkflowPollers) > 0 || len(n.ActivityPollers) > 0)
}
