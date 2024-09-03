package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) created(ctx workflow.Context) error {
	// our logic goes here
	w.startReconcileEventLoopsWorkflowCron(ctx, activities.EnsureEventLoopsRequest{})
	return nil
}
