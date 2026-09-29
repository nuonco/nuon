package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type WorkflowStepGenerator func(ctx workflow.Context, uf *app.Workflow) (*app.GenerateStepsResult, error)
