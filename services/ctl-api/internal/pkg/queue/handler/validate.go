package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (h *handler) validateHandler(ctx workflow.Context) (*ValidateResponse, error) {
	if h.sig == nil {
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
			return nil, panicErr
		}

		validateErr := &signal.SignalErrValidate{Err: err}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: validateErr.Error(),
			Metadata: map[string]any{
				"validate_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		return nil, validateErr
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
			retErr = &signal.SignalErrPanic{Value: r, Phase: "validate"}
		}
	}()

	return h.sig.Validate(ctx)
}
