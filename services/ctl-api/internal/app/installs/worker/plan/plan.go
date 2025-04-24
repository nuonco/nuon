package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateActionRunPlanRequest struct {
	ActionWorkflowRunID string `validate:"required"`

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback WorkflowIDCallback
func CreateActionWorkflowRunPlan(ctx workflow.Context, req *CreateActionRunPlanRequest) (*plantypes.ActionWorkflowRunPlan, error) {
	p := Planner{}
	return p.createActionWorkflowRunPlan(ctx, req.ActionWorkflowRunID)
}
