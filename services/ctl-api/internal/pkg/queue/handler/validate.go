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
	defer func() {
		status := "success"
		desc := ""
		if retErr != nil {
			status = "error"
			desc = retErr.Error()
		}
		callback.Send(ctx, l, cb, callback.Result{Status: status, StatusDescription: desc})
	}()

	if err := workflow.Await(ctx, func() bool {
		return h.ready
	}); err != nil {
		h.setFinished(app.StatusError, err.Error())
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	if h.sig == nil {
		h.setFinished(app.StatusError, "signal was empty can not proceed")
		return nil, errors.New("signal was empty can not proceed")
	}

	// mark the signal as in-progress in the DB
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusInProgress,
		Metadata: map[string]any{
			"validate_started_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
		},
	})

	event := h.buildSignalPhaseEvent(signal.SignalPhaseValidate)

	// run before-phase hooks (fail-open)
	decision := h.runBeforePhase(ctx, event)
	if !decision.Allow {
		blockedErr := &signal.SignalErrValidate{Err: errors.New("blocked by lifecycle hook: " + decision.Reason)}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: blockedErr.Error(),
			Metadata: map[string]any{
				"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		h.setFinished(app.StatusError, blockedErr.Error())
		return nil, blockedErr
	}

	start := workflow.Now(ctx)
	err := h.runSignalValidate(ctx)
	dur := workflow.Now(ctx).Sub(start)

	// run after-phase hooks (best-effort)
	h.runAfterPhaseSafe(ctx, event, outcomeFromError(err, dur))

	if err != nil {
		// If the signal panicked, write error status here (outside the panic boundary).
		var panicErr *signal.SignalErrPanic
		if errors.As(err, &panicErr) {
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: panicErr.Error(),
				Metadata: map[string]any{
					"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
				},
			})
			h.setFinished(app.StatusError, panicErr.Error())
			return nil, panicErr
		}

		validateErr := &signal.SignalErrValidate{Err: err}
		humanDesc := signal.HumanError(err)
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: humanDesc,
			Metadata: map[string]any{
				"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		h.setFinished(app.StatusError, humanDesc)
		return nil, temporal.NewNonRetryableApplicationError(
			"signal failure",
			humanDesc,
			validateErr)
	}

	// record validate completion timestamp
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusInProgress,
		Metadata: map[string]any{
			"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
		},
	})

	return nil, nil
}

// runSignalValidate calls the user-provided signal Validate in a panic-safe boundary.
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
