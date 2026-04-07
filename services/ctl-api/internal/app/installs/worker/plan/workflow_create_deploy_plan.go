package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

type CreateDeployPlanRequest struct {
	InstallDeployID string
	InstallID       string

	WorkflowID string
}

type CreateDeployPlanResponse struct {
	Plan          *plantypes.DeployPlan
	RoleSelection *operationroles.RoleSelection
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-template {{.Req.WorkflowID}}
func CreateDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (CreateDeployPlanResponse, error) {
	p := Planner{}
	plan, roleSelection, err := p.createDeployPlan(ctx, req)
	return CreateDeployPlanResponse{
		Plan:          plan,
		RoleSelection: roleSelection,
	}, err
}
