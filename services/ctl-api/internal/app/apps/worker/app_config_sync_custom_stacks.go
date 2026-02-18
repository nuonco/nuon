package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
)

type AppConfigSyncCustomStacksRequest struct {
	OrgID            string `validate:"required"`
	AppID            string `validate:"required"`
	AppStackConfigID string `validate:"required"`
}

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 1m
// @id-template app-config-sync-custom-stacks-{{.AppStackConfigID}}
func (w *Workflows) AppConfigSyncCustomStacks(ctx workflow.Context, req AppConfigSyncCustomStacksRequest) error {
	return activities.AwaitUploadCustomNestedStackTemplates(ctx, &activities.UploadCustomNestedStackTemplatesRequest{
		AppStackConfigID: req.AppStackConfigID,
	})
}
