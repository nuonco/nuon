package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

type CreateComponentBuildPlanRequest struct {
	ComponentID      string
	ComponentBuildID string

	WorkflowID           string
	CloudProvider        string
	ManagementIAMRoleARN string
	IsControlPlaneBuild  bool
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-template {{.Req.WorkflowID}}
func CreateComponentBuildPlan(ctx workflow.Context, req *CreateComponentBuildPlanRequest) (*plantypes.BuildPlan, error) {
	p := Planner{
		cloudProvider:        req.CloudProvider,
		managementIAMRoleARN: req.ManagementIAMRoleARN,
		isControlPlaneBuild:  req.IsControlPlaneBuild,
	}
	return p.createComponentBuildPlan(ctx, req)
}
