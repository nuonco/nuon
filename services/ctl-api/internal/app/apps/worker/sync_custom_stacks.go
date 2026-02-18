package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
)

// @temporal-gen workflow
func (w *Workflows) SyncCustomStacks(ctx workflow.Context, sreq signals.RequestSignal) error {
	return activities.AwaitUploadCustomNestedStackTemplates(ctx, &activities.UploadCustomNestedStackTemplatesRequest{
		AppStackConfigID: sreq.AppStackConfigID,
	})
}
