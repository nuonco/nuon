package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateSyncPlanRequest struct {
	InstallDeployID string
	InstallID       string

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback CreateSyncWorkflowIDCallback
func CreateSyncPlan(ctx workflow.Context, req *CreateSyncPlanRequest) (*plantypes.SyncOCIPlan, error) {
	p := Planner{}
	return p.createSyncPlan(ctx, req)
}
