package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

type CreateSyncSecretsPlanRequest struct {
	InstallID string

	WorkflowID string
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-callback SyncSecretsWorkflowIDCallback
func CreateSyncSecretsPlan(ctx workflow.Context, req *CreateSyncSecretsPlanRequest) (*plantypes.SyncSecretsPlan, error) {
	p := Planner{}
	return p.createSyncSecretsPlan(ctx, req)
}
