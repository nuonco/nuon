package worker

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
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

	// NOTE(jm): this is going to be replaced with a way to coordinate signals across namespaces/concurrency models,
	// but for now we're doing everything in this one process.
	var sig signals.Signal
	if err := json.Unmarshal(step.Signal.SignalJSON, &sig); err != nil {
		return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	sig.WorkflowStepID = step.ID

	handlers := w.getHandlers()
	handler, ok := handlers[sig.Type]
	if !ok {
		err := errors.New(fmt.Sprintf("no handler found for signal %s", sig.Type))
		return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	// now, execute the handler
	handlerSreq := signals.NewRequestSignal(sreq.EventLoopRequest, &sig)
	if err := handler(ctx, handlerSreq); err != nil {
		return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
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
