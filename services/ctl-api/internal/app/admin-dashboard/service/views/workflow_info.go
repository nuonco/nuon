package views

import (
	"time"
)

// WorkflowInfo holds Temporal workflow execution info for display.
type WorkflowInfo struct {
	Status     string
	Activities []ActivityInfo
}

// ActivityInfo holds Temporal activity execution info for display.
type ActivityInfo struct {
	Name       string
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Attempt    int32
	Failure    string
	Input      string // JSON-formatted activity input
	Result     string // JSON-formatted activity result
}
