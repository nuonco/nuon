package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (h *handler) validateHandler(ctx workflow.Context) (resp *ValidateResponse, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := &signal.SignalErrPanic{Value: r, Phase: "validate"}
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
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
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusInProgress,
	})

	hooks := h.sig.GetHooks()

	result, _ := hooks.PreExecuteHooks(ctx, signal.SignalPhaseValidate)
	if !result.Allow {
		validateErr := &signal.SignalErrValidate{Err: errors.New((&signal.SignalErrBlocked{Phase: signal.SignalPhaseValidate, Reason: result.Reason}).Error())}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: validateErr.Error(),
		})
		return nil, validateErr
	}

	start := workflow.Now(ctx)
	err := h.sig.Validate(ctx)
	dur := workflow.Now(ctx).Sub(start)

	hooks.PostExecuteHooks(ctx, result.Event, hooks.BuildOutcome(err, dur))

	if err != nil {
		validateErr := &signal.SignalErrValidate{Err: err}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: validateErr.Error(),
		})
		return nil, validateErr
	}

	return nil, nil
}
