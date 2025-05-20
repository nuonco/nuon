package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateComponentBuildPlanRequest struct {
	ComponentID      string
	ComponentBuildID string

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback CreateComponentBuildWorkflowIDCallback
func CreateComponentBuildPlan(ctx workflow.Context, req *CreateComponentBuildPlanRequest) (*plantypes.BuildPlan, error) {
	p := Planner{}
	return p.createComponentBuildPlan(ctx, req)
}
