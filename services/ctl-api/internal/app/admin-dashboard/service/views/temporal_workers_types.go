package views

import (
	"time"
)

// NamespaceWorkerInfo holds worker/poller information for a single Temporal namespace.
type NamespaceWorkerInfo struct {
	Namespace string `json:"namespace"`
	TaskQueue string `json:"task_queue"`
	Error     string `json:"error,omitempty"` // non-empty if we failed to query this namespace

	WorkflowPollers []PollerDetail      `json:"workflow_pollers"`
	ActivityPollers []PollerDetail      `json:"activity_pollers"`
	WorkflowStats   *TaskQueueStatsInfo `json:"workflow_stats,omitempty"`
	ActivityStats   *TaskQueueStatsInfo `json:"activity_stats,omitempty"`
}

// PollerDetail represents a single worker/poller connected to a task queue.
type PollerDetail struct {
	Identity       string    `json:"identity"`
	LastAccessTime time.Time `json:"last_access_time"`
	RatePerSecond  float64   `json:"rate_per_second"`
}

// TaskQueueStatsInfo holds stats about a task queue.
type TaskQueueStatsInfo struct {
	ApproximateBacklogCount int64         `json:"approximate_backlog_count"`
	ApproximateBacklogAge   time.Duration `json:"approximate_backlog_age"`
	TasksAddRate            float32       `json:"tasks_add_rate"`
	TasksDispatchRate       float32       `json:"tasks_dispatch_rate"`
}

// TotalPollerCount returns the total number of pollers across workflow and activity queues.
func (n *NamespaceWorkerInfo) TotalPollerCount() int {
	return len(n.WorkflowPollers) + len(n.ActivityPollers)
}

// IsHealthy returns true if there are active pollers and no error.
func (n *NamespaceWorkerInfo) IsHealthy() bool {
	return n.Error == "" && (len(n.WorkflowPollers) > 0 || len(n.ActivityPollers) > 0)
}
