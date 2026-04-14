package views

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// WorkflowInfo holds Temporal workflow execution info for display.
type WorkflowInfo struct {
	Status           string
	Activities       []ActivityInfo
	ChildWorkflows   []ChildWorkflowInfo
	AwaitedSignals   []AwaitedSignalInfo
	UpdateHandlers   []string
	UpdateExecutions []UpdateExecution
	OrphanActivities []ActivityInfo // activities not in any update (main workflow body)
}

// UpdateExecution groups activities that ran within a single Temporal update handler call.
type UpdateExecution struct {
	Name       string // update handler name (e.g. "execute", "retry-step")
	UpdateID   string
	Status     string // Accepted, Completed, Failed, Rejected, Running
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Input      string // JSON-formatted update input args
	Result     string // JSON-formatted update result
	Failure    string
	Activities []ActivityInfo
}

// ActivityInfo holds Temporal activity execution info for display.
type ActivityInfo struct {
	Name             string
	Status           string
	StartedAt        time.Time
	FinishedAt       time.Time
	Duration         time.Duration
	Attempt          int32
	Failure          string
	Input            string // JSON-formatted activity input
	Result           string // JSON-formatted activity result
	ScheduledEventID int64  // Temporal event ID for correlating with updates
}

// ChildWorkflowInfo holds info about a child workflow execution.
type ChildWorkflowInfo struct {
	WorkflowType string
	WorkflowID   string
	RunID        string
	Namespace    string
	Status       string
	StartedAt    time.Time
	FinishedAt   time.Time
	Duration     time.Duration
	Failure      string
}

// AwaitedSignalInfo holds info about a queue signal that was awaited.
type AwaitedSignalInfo struct {
	QueueSignalID string
	Signal        *app.QueueSignal // loaded from DB if available
	Status        string           // activity status: Completed, Failed, Running
	StartedAt     time.Time
	FinishedAt    time.Time
	Duration      time.Duration
	Failure       string
}
