package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (h *handler) validateHandler(ctx workflow.Context) (resp *ValidateResponse, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := &signal.SignalErrPanic{Value: r, Phase: "validate"}
			_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: panicErr.Error(),
			})
			retErr = panicErr
		}
	}()

	if h.sig == nil {
		return nil, errors.New("signal was empty can not proceed")
	}

	// mark the signal as in-progress in the DB
	_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusInProgress,
	})

	event := h.buildSignalPhaseEvent(signal.SignalPhaseValidate)

	// run before-phase hooks (fail-open)
	decision := h.runBeforePhase(ctx, event)
	if !decision.Allow {
		blockedErr := &signal.SignalErrValidate{Err: errors.New("blocked by lifecycle hook: " + decision.Reason)}
		_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: blockedErr.Error(),
		})
		return nil, blockedErr
	}

	start := workflow.Now(ctx)
	err := h.sig.Validate(ctx)
	dur := workflow.Now(ctx).Sub(start)

	// run after-phase hooks (best-effort)
	h.runAfterPhaseSafe(ctx, event, outcomeFromError(err, dur))

	if err != nil {
		validateErr := &signal.SignalErrValidate{Err: err}
		_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: validateErr.Error(),
		})
		return nil, validateErr
	}

	return nil, nil
}
