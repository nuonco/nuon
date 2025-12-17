package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker/activities"
	installsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/pkg/errors"
)

// @temporal-gen workflow
// @execution-timeout 5m
// @task-timeout 3m
func (w *Workflows) Delete(ctx workflow.Context, sreq signals.RequestSignal) error {
	acw, err := activities.AwaitGetActionWorkflowByWorkflowID(ctx, sreq.ID)
	if err != nil {
		return err
	}

	if acw == nil {
		return nil
	}

	err = activities.AwaitDeleteActionWorkflowByActionWorkflowID(ctx, sreq.ID)
	if err != nil {
		if statusErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			WorkflowID:        sreq.ID,
			Status:            app.ActionWorkflowStatusError,
			StatusDescription: "unable to delete action workflow",
		}); statusErr != nil {
			return errors.Wrap(statusErr, "unable to update status")
		}

		return errors.Wrap(err, "unable to delete action workflow")
	}

	installIDs, err := activities.AwaitGetActionWorkflowInstallsByActionWorkflowID(ctx, sreq.ID)
	if err != nil {
		if statusErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			WorkflowID:        sreq.ID,
			Status:            app.ActionWorkflowStatusError,
			StatusDescription: "unable to delete action workflow",
		}); statusErr != nil {
			return errors.Wrap(statusErr, "unable to update status")
		}

		return errors.Wrap(err, "unable to get action workflow installs")
	}

	for _, installID := range installIDs {
		w.evClient.Send(ctx, installID, &installsignals.Signal{
			Type: installsignals.OperationRestart,
		})
	}

	return nil
}
