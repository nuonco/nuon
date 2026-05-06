package views

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// StepDetailData holds enriched step data for the workflow detail view.
type StepDetailData struct {
	Step              *app.WorkflowStep `json:"step"`
	QueueSignalJSON   string            `json:"queue_signal_json,omitempty"`
	StepSignalID      string            `json:"step_signal_id,omitempty"`
	StepSignalQueueID string            `json:"step_signal_queue_id,omitempty"`
	StepTarget        *StepTargetData   `json:"step_target,omitempty"`
}

// GroupDetailData holds a step group and its enriched steps for the workflow detail view.
type GroupDetailData struct {
	Group *app.WorkflowStepGroup `json:"group"`
	Steps []StepDetailData       `json:"steps"`
}

// StepTargetData holds the loaded step target with its log stream.
type StepTargetData struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}
