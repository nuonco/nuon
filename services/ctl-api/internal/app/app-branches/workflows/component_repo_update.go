package workflows

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func ComponentRepoUpdateSteps(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	return steps, nil
}
