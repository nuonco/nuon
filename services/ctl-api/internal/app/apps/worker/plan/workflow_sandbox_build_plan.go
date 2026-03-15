package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

type CreateSandboxBuildPlanRequest struct {
	AppSandboxBuildID string
	WorkflowID        string
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
func CreateSandboxBuildPlan(ctx workflow.Context, req *CreateSandboxBuildPlanRequest) (*plantypes.BuildPlan, error) {
	p := Planner{}
	return p.createSandboxBuildPlan(ctx, req)
}
