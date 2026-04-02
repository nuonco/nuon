package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

type CreateSandboxRunPlanRequest struct {
	RunID      string
	InstallID  string
	RootDomain string

	WorkflowID string
}

type CreateSandboxPlanResponse struct {
	Plan          *plantypes.SandboxRunPlan
	RoleSelection *operationroles.RoleSelection
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-template {{.Req.WorkflowID}}
func CreateSandboxRunPlan(ctx workflow.Context, req *CreateSandboxRunPlanRequest) (CreateSandboxPlanResponse, error) {
	p := Planner{}
	plan, roleSelection, err := p.createSandboxRunPlan(ctx, req)

	return CreateSandboxPlanResponse{
		Plan:          plan,
		RoleSelection: roleSelection,
	}, err
}
