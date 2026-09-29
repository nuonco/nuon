package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (h *handler) validateHandler(ctx workflow.Context, cb callback.Ref) (resp *ValidateResponse, retErr error) {
	l, _ := log.WorkflowLogger(ctx)
	h.validating = true

	// why: Apply the terminal status only after the completion callback has been
	// sent. Setting h.finished earlier lets run() complete the workflow and
	// abandon the in-flight callback activity, which drops the callback the
	// dispatcher is waiting on and wedges the queue.
	var finStatus app.Status
	var finDesc string
	defer func() {
		status := "success"
		desc := ""
		switch {
		case finStatus == app.StatusCancelled:
			status = "cancelled"
			desc = finDesc
		case retErr != nil:
			status = "error"
			desc = retErr.Error()
		}
		callback.Send(ctx, l, cb, callback.Result{Status: status, StatusDescription: desc})
		if finStatus != "" {
			h.setFinished(finStatus, finDesc)
		}
		h.validating = false
	}()

	if h.canceled {
		finStatus, finDesc = app.StatusCancelled, "signal was canceled"
		return nil, errors.New("signal was canceled")
	}

	if err := workflow.Await(ctx, func() bool {
		return h.ready
	}); err != nil {
		finStatus, finDesc = app.StatusError, err.Error()
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	if h.sig == nil {
		finStatus, finDesc = app.StatusError, "signal was empty can not proceed"
		return nil, errors.New("signal was empty can not proceed")
	}

	var err error
	ctx, err = h.signalContext(ctx, true)
	if err != nil {
		finStatus, finDesc = app.StatusError, err.Error()
		return nil, errors.Wrap(err, "unable to restore signal context")
	}
	l, _ = log.WorkflowLogger(ctx)

	if !h.skipValidateStamps() {
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID: h.queueSignalID,
			Status:        app.StatusInProgress,
			Metadata: map[string]any{
				"validate_started_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
	}

	event := h.buildSignalPhaseEvent(signal.SignalPhaseValidate)

	decision := h.runBeforePhase(ctx, event)
	if !decision.Allow {
		blockedErr := &signal.SignalErrValidate{Err: errors.New("blocked by lifecycle hook: " + decision.Reason)}
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: blockedErr.Error(),
			Metadata: map[string]any{
				"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		finStatus, finDesc = app.StatusError, blockedErr.Error()
		return nil, blockedErr
	}

	start := workflow.Now(ctx)
	err = h.runSignalValidate(ctx)
	dur := workflow.Now(ctx).Sub(start)

	h.runAfterPhaseSafe(ctx, event, outcomeFromError(err, dur))

	if err != nil {
		var panicErr *signal.SignalErrPanic
		if errors.As(err, &panicErr) {
			_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: panicErr.Error(),
				Metadata: map[string]any{
					"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
				},
			})
			finStatus, finDesc = app.StatusError, panicErr.Error()
			return nil, panicErr
		}

		validateErr := &signal.SignalErrValidate{Err: err}
		humanDesc := signal.HumanError(err)
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: humanDesc,
			Metadata: map[string]any{
				"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		finStatus, finDesc = app.StatusError, humanDesc
		return nil, temporal.NewNonRetryableApplicationError(
			"signal failure",
			humanDesc,
			validateErr)
	}

	skipCancelledStamp := h.canceled &&
		workflow.GetVersion(ctx, handlerCancelledStatusVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	if !h.skipValidateStamps() && !skipCancelledStamp {
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID: h.queueSignalID,
			Status:        app.StatusInProgress,
			Metadata: map[string]any{
				"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
	}

	return nil, nil
}

func (h *handler) skipValidateStamps() bool {
	iv, ok := h.sig.(signal.SignalWithInlineValidate)
	return ok && iv.InlineValidate()
}

func (h *handler) runSignalValidate(ctx workflow.Context) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = signal.NewSignalErrPanic(r, "validate")
		}
	}()

	sig, err := h.checkSandboxMode(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to check sandbox mode")
	}

	return sig.Validate(ctx)
}
