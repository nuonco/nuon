package worker

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

func ExecuteWorkflowStepWorkflowID(req signals.RequestSignal) string {
	return fmt.Sprintf("%s-execute-workflow-step-%s", req.WorkflowID(req.ID), req.WorkflowStepID)
}

// @id-callback ExecuteWorkflowStepWorkflowID
// @temporal-gen workflow
// @execution-timeout 120m
func (w *Workflows) ExecuteWorkflowStep(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	defer func() {
		// NOTE(jm): this should be a helper function.
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()

			if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: sreq.WorkflowStepID,
				Status: app.CompositeStatus{
					Status: app.StatusCancelled,
				},
			}); err != nil {
				l.Error("unable to update step status on cancellation", zap.Error(err))
			}
		}
	}()

	if err := activities.AwaitUpdateWorkflowStepStartedAtByID(ctx, sreq.WorkflowStepID); err != nil {
		return err
	}
	defer func() {
		if err := activities.AwaitUpdateWorkflowStepFinishedAtByID(ctx, sreq.WorkflowStepID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: sreq.WorkflowStepID,
		Status: app.CompositeStatus{
			Status: app.StatusInProgress,
		},
	}); err != nil {
		return err
	}

	// build the signal here
	step, err := activities.AwaitGetInstallWorkflowsStepByInstallWorkflowStepID(ctx, sreq.WorkflowStepID)
	if err != nil {
		return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	if step.ExecutionType == app.InstallWorkflowStepExecutionTypeSkipped {
		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: sreq.WorkflowStepID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	var sig signals.Signal
	if err := json.Unmarshal(step.Signal.SignalJSON, &sig); err != nil {
		return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	ctx = cctx.SetInstallWorkflowWorkflowContext(ctx, &app.InstallWorkflowContext{
		ID:             step.InstallWorkflowID,
		WorkflowStepID: &step.ID,
	})
	sig.WorkflowStepID = step.ID
	sig.WorkflowStepName = step.Name
	handlerSreq := signals.NewRequestSignal(sreq.EventLoopRequest, &sig)

	// Predict the workflow ID of the underlying object's event loop
	suffix, err := w.subloopSuffix(ctx, handlerSreq)
	if err != nil {
		return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	if suffix != "" {
		id := fmt.Sprintf("%s-%s", sreq.ID, suffix)
		if err := w.evClient.SendAndWait(ctx, id, &handlerSreq); err != nil {
			return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
		}
	} else {
		// no suffix means a workflow on this loop, so we must invoke the handler directly
		handlers := w.getHandlers()
		handler, ok := handlers[sig.Type]
		if !ok {
			err := errors.New(fmt.Sprintf("no handler found for signal %s", sig.Type))
			return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
		}
		if err := handler(ctx, handlerSreq); err != nil {
			return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
		}
	}

	// update the status at the end
	if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: sreq.WorkflowStepID,
		Status: app.CompositeStatus{
			Status: app.StatusSuccess,
		},
	}); err != nil {
		return err
	}
	return nil
}

func (w *Workflows) handleStepErr(ctx workflow.Context, installWorkflowStepID string, err error) error {
	if statusErr := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: installWorkflowStepID,
		Status: app.CompositeStatus{
			Status: app.StatusError,
			Metadata: map[string]any{
				"err_message": err.Error(),
			},
		},
	}); statusErr != nil {
		return status.WrapStatusErr(err, statusErr)
	}

	return err
}
