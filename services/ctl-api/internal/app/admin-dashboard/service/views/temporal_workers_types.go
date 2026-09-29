package views

import (
	"time"
)

type NamespaceWorkerInfo struct {
	Namespace string `json:"namespace"`
	TaskQueue string `json:"task_queue"`
	Error     string `json:"error,omitempty"`

	WorkflowPollers []PollerDetail      `json:"workflow_pollers"`
	ActivityPollers []PollerDetail      `json:"activity_pollers"`
	WorkflowStats   *TaskQueueStatsInfo `json:"workflow_stats,omitempty"`
	ActivityStats   *TaskQueueStatsInfo `json:"activity_stats,omitempty"`
}

type PollerDetail struct {
	Identity       string    `json:"identity"`
	LastAccessTime time.Time `json:"last_access_time"`
	RatePerSecond  float64   `json:"rate_per_second"`
}

type TaskQueueStatsInfo struct {
	ApproximateBacklogCount int64         `json:"approximate_backlog_count"`
	ApproximateBacklogAge   time.Duration `json:"approximate_backlog_age"`
	TasksAddRate            float32       `json:"tasks_add_rate"`
	TasksDispatchRate       float32       `json:"tasks_dispatch_rate"`
}

func (n *NamespaceWorkerInfo) TotalPollerCount() int {
	return len(n.WorkflowPollers) + len(n.ActivityPollers)
}

func (n *NamespaceWorkerInfo) IsHealthy() bool {
	return n.Error == "" && (len(n.WorkflowPollers) > 0 || len(n.ActivityPollers) > 0)
}
