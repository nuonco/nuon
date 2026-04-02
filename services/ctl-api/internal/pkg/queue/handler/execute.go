package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const ExecuteUpdateName string = "execute"

const executeUpdateType = handlerTypeUpdate

type ExecuteResponse struct{}

func (h *handler) executeHandler(ctx workflow.Context) (resp *ExecuteResponse, retErr error) {
	defer func() {
		h.finished = true
		h.executingCtx = nil
		h.executingCancel = nil
	}()

	defer func() {
		if r := recover(); r != nil {
			panicErr := &signal.SignalErrPanic{Value: r, Phase: "execute"}
			_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: panicErr.Error(),
			})
			retErr = panicErr
		}
	}()

	if h.canceled {
		return nil, errors.New("signal was canceled")
	}

	execCtx, cancel := workflow.WithCancel(ctx)
	h.executingCtx = execCtx
	h.executingCancel = cancel
	defer cancel()

	if err := h.sig.Execute(execCtx); err != nil {
		if h.canceled {
			// canceled mid-execute — cancelHandler already wrote StatusCancelled
			return nil, errors.Wrap(err, "signal was canceled during execution")
		}

		execErr := &signal.SignalErrExecute{Err: err}
		_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: execErr.Error(),
		})
		return nil, execErr
	}

	// persist success status to DB
	_ = activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusSuccess,
	})

	return nil, nil
}
