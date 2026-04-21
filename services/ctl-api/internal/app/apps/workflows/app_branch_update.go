package workflows

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"
)

func AppBranchUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	// flow step definition goes here
	return nil, nil
}
