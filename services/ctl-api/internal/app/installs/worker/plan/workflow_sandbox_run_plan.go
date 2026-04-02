package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

type CreateSandboxRunPlanRequest struct {
	RunID      string
	InstallID  string
	RootDomain string

	WorkflowID string
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-template {{.Req.WorkflowID}}
func CreateSandboxRunPlan(ctx workflow.Context, req *CreateSandboxRunPlanRequest) (*plantypes.SandboxRunPlan, error) {
	p := Planner{}
	return p.createSandboxRunPlan(ctx, req)
}
