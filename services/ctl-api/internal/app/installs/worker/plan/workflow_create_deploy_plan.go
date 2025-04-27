package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateDeployPlanRequest struct {
	InstallDeployID string
	InstallID       string

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback CreateDeployPlanIDCallback
func CreateDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.DeployPlan, error) {
	p := Planner{}

	return p.createDeployPlan(ctx, req)
}
