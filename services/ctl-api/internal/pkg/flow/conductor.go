package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// WorkflowStepGenerator is a function that generates workflow steps for a given workflow.
type WorkflowStepGenerator func(ctx workflow.Context, uf *app.Workflow) (*app.GenerateStepsResult, error)
