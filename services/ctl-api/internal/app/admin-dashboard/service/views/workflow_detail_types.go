package views

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type StepDetailData struct {
	Step              *app.WorkflowStep `json:"step"`
	QueueSignalJSON   string            `json:"queue_signal_json,omitempty"`
	StepSignalID      string            `json:"step_signal_id,omitempty"`
	StepSignalQueueID string            `json:"step_signal_queue_id,omitempty"`
	StepTarget        *StepTargetData   `json:"step_target,omitempty"`
}

type GroupDetailData struct {
	Group *app.WorkflowStepGroup `json:"group"`
	Steps []StepDetailData       `json:"steps"`
}

type StepTargetData struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}
