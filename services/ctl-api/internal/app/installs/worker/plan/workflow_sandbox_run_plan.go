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

	// RoleARN is the pre-selected IAM role ARN to use for the plan's auth
	// (both AWSAuth and Hooks.RunAuth). When set, it overrides the default
	// derivation from stack outputs in getAuth. This is required when the
	// default operation role (provision/deprovision) is disabled via
	// EnableRunner* CloudFormation parameters.
	RoleARN string
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
