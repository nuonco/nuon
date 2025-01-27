package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateActionRunPlanRequest struct {
	RunID string

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback WorkflowIDCallback
func CreateActionWorkflowRunPlan(ctx workflow.Context, req *CreateActionRunPlanRequest) (*plantypes.ActionWorkflowRunPlan, error) {
	p := planner{}
	return p.createPlan(ctx, req.RunID)
}
