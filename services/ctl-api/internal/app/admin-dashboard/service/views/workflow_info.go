package views

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// WorkflowInfo holds Temporal workflow execution info for display.
type WorkflowInfo struct {
	Status           string               `json:"status"`
	Activities       []ActivityInfo       `json:"activities"`
	ChildWorkflows   []ChildWorkflowInfo  `json:"child_workflows"`
	AwaitedSignals   []AwaitedSignalInfo  `json:"awaited_signals"`
	EnqueuedSignals  []EnqueuedSignalInfo `json:"enqueued_signals"`
	UpdateHandlers   []string             `json:"update_handlers"`
	UpdateExecutions []UpdateExecution    `json:"update_executions"`
	OrphanActivities []ActivityInfo       `json:"orphan_activities"` // activities not in any update (main workflow body)
}

// EnqueuedSignalInfo holds info about a signal that was enqueued by this workflow.
type EnqueuedSignalInfo struct {
	QueueSignalID string           `json:"queue_signal_id"`
	Signal        *app.QueueSignal `json:"signal"`
	ActivityName  string           `json:"activity_name"`
}

// UpdateExecution groups activities that ran within a single Temporal update handler call.
type UpdateExecution struct {
	Name            string               `json:"name"` // update handler name (e.g. "execute", "retry-step")
	UpdateID        string               `json:"update_id"`
	Status          string               `json:"status"` // Accepted, Completed, Failed, Rejected, Running
	StartedAt       time.Time            `json:"started_at"`
	FinishedAt      time.Time            `json:"finished_at"`
	Duration        time.Duration        `json:"duration"`
	Input           string               `json:"input"`  // JSON-formatted update input args
	Result          string               `json:"result"` // JSON-formatted update result
	Failure         string               `json:"failure"`
	Activities      []ActivityInfo       `json:"activities"`
	AwaitedSignals  []AwaitedSignalInfo  `json:"awaited_signals"`
	EnqueuedSignals []EnqueuedSignalInfo `json:"enqueued_signals"`
}

// ActivityInfo holds Temporal activity execution info for display.
type ActivityInfo struct {
	Name             string        `json:"name"`
	Status           string        `json:"status"`
	StartedAt        time.Time     `json:"started_at"`
	FinishedAt       time.Time     `json:"finished_at"`
	Duration         time.Duration `json:"duration"`
	Attempt          int32         `json:"attempt"`
	Failure          string        `json:"failure"`
	Input            string        `json:"input"`              // JSON-formatted activity input
	Result           string        `json:"result"`             // JSON-formatted activity result
	ScheduledEventID int64         `json:"scheduled_event_id"` // Temporal event ID for correlating with updates
}

// ChildWorkflowInfo holds info about a child workflow execution.
type ChildWorkflowInfo struct {
	WorkflowType string        `json:"workflow_type"`
	WorkflowID   string        `json:"workflow_id"`
	RunID        string        `json:"run_id"`
	Namespace    string        `json:"namespace"`
	Status       string        `json:"status"`
	StartedAt    time.Time     `json:"started_at"`
	FinishedAt   time.Time     `json:"finished_at"`
	Duration     time.Duration `json:"duration"`
	Failure      string        `json:"failure"`
}

// AwaitedSignalInfo holds info about a queue signal that was awaited.
type AwaitedSignalInfo struct {
	QueueSignalID string           `json:"queue_signal_id"`
	Signal        *app.QueueSignal `json:"signal"` // loaded from DB if available
	Status        string           `json:"status"` // activity status: Completed, Failed, Running
	StartedAt     time.Time        `json:"started_at"`
	FinishedAt    time.Time        `json:"finished_at"`
	Duration      time.Duration    `json:"duration"`
	Failure       string           `json:"failure"`
}
