package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

// @temporal-gen workflow
// @execution-timeout 120m
func (w *Workflows) ExecuteWorkflow(ctx workflow.Context, sreq signals.RequestSignal) error {
	// generate steps
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	if err := activities.AwaitUpdateWorkflowStartedAtByID(ctx, sreq.InstallWorkflowID); err != nil {
		return err
	}
	defer func() {
		if err := activities.AwaitUpdateWorkflowFinishedAtByID(ctx, sreq.InstallWorkflowID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	l.Debug("generating steps for workflow")
	if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: sreq.InstallWorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "generating steps for workflow",
		},
	}); err != nil {
		return err
	}
	if err := w.AwaitGenerateWorkflowSteps(ctx, sreq); err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: sreq.InstallWorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "error while generating steps",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to generate workflow steps")
	}
	if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: sreq.InstallWorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "successfully generated all steps",
		},
	}); err != nil {
		return err
	}

	l.Debug("executing steps for workflow")
	if err := w.AwaitExecuteWorkflowSteps(ctx, sreq); err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: sreq.InstallWorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "error while executing steps",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to execute workflow steps")
	}

	if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: sreq.InstallWorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "successfully executed install workflow",
		},
	}); err != nil {
		return err
	}

	return nil
}
