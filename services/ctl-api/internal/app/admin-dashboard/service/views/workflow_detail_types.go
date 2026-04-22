package views

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// StepDetailData holds enriched step data for the workflow detail view.
type StepDetailData struct {
	Step            *app.WorkflowStep
	QueueSignalJSON string
	StepTarget      *StepTargetData
}

// GroupDetailData holds a step group and its enriched steps for the workflow detail view.
type GroupDetailData struct {
	Group *app.WorkflowStepGroup
	Steps []StepDetailData
}

// StepTargetData holds the loaded step target with its log stream.
type StepTargetData struct {
	ID          string
	Type        string
	Status      string
	LogStreamID string
}
