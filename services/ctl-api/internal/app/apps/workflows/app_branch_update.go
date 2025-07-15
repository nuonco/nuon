package workflows

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"
)

func AppBranchUpdate(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	// flow step definition goes here
	return nil, nil
}
