package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

type CreateActionRunPlanRequest struct {
	ActionWorkflowRunID string `validate:"required"`

	WorkflowID string
}

type CreateActionPlanResponse struct {
	Plan          *plantypes.ActionWorkflowRunPlan
	RoleSelection *operationroles.RoleSelection
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-template {{.Req.WorkflowID}}
func CreateActionWorkflowRunPlan(ctx workflow.Context, req *CreateActionRunPlanRequest) (CreateActionPlanResponse, error) {
	p := Planner{}
	plan, roleSelection, err := p.createActionWorkflowRunPlan(ctx, req.ActionWorkflowRunID)
	return CreateActionPlanResponse{
		Plan:          plan,
		RoleSelection: roleSelection,
	}, err
}
