package views

import (
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type WorkflowInfo struct {
	Status           string               `json:"status"`
	Activities       []ActivityInfo       `json:"activities"`
	ChildWorkflows   []ChildWorkflowInfo  `json:"child_workflows"`
	AwaitedSignals   []AwaitedSignalInfo  `json:"awaited_signals"`
	EnqueuedSignals  []EnqueuedSignalInfo `json:"enqueued_signals"`
	UpdateHandlers   []string             `json:"update_handlers"`
	UpdateExecutions []UpdateExecution    `json:"update_executions"`
	OrphanActivities []ActivityInfo       `json:"orphan_activities"`
}

type EnqueuedSignalInfo struct {
	QueueSignalID string           `json:"queue_signal_id"`
	Signal        *app.QueueSignal `json:"signal"`
	ActivityName  string           `json:"activity_name"`
}

type UpdateExecution struct {
	Name            string               `json:"name"`
	UpdateID        string               `json:"update_id"`
	Status          string               `json:"status"`
	StartedAt       time.Time            `json:"started_at"`
	FinishedAt      time.Time            `json:"finished_at"`
	Duration        time.Duration        `json:"duration"`
	Input           string               `json:"input"`
	Result          string               `json:"result"`
	Failure         string               `json:"failure"`
	Activities      []ActivityInfo       `json:"activities"`
	AwaitedSignals  []AwaitedSignalInfo  `json:"awaited_signals"`
	EnqueuedSignals []EnqueuedSignalInfo `json:"enqueued_signals"`
}

type ActivityInfo struct {
	Name             string        `json:"name"`
	Status           string        `json:"status"`
	StartedAt        time.Time     `json:"started_at"`
	FinishedAt       time.Time     `json:"finished_at"`
	Duration         time.Duration `json:"duration"`
	Attempt          int32         `json:"attempt"`
	Failure          string        `json:"failure"`
	Input            string        `json:"input"`
	Result           string        `json:"result"`
	ScheduledEventID int64         `json:"scheduled_event_id"`
}

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

type AwaitedSignalInfo struct {
	QueueSignalID string           `json:"queue_signal_id"`
	Signal        *app.QueueSignal `json:"signal"`
	Status        string           `json:"status"`
	StartedAt     time.Time        `json:"started_at"`
	FinishedAt    time.Time        `json:"finished_at"`
	Duration      time.Duration    `json:"duration"`
	Failure       string           `json:"failure"`
}
