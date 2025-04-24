package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateSandboxRunPlanRequest struct {
	RunID     string
	InstallID string

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback SandboxRunWorkflowIDCallback
func CreateSandboxRunPlan(ctx workflow.Context, req *CreateSandboxRunPlanRequest) (*plantypes.SandboxRunPlan, error) {
	p := Planner{}
	return p.createSandboxRunPlan(ctx, req)
}
