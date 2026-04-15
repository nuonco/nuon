package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
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
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: panicErr.Error(),
			})
			retErr = panicErr
		}
	}()

	// Increment execution count to track how many times this signal has been executed.
	_ = activities.AwaitIncrementQueueSignalExecutionCount(ctx, &activities.IncrementQueueSignalExecutionCountRequest{
		QueueSignalID: h.queueSignalID,
	})

	if h.canceled {
		return nil, errors.New("signal was canceled")
	}

	execCtx, cancel := workflow.WithCancel(ctx)
	h.executingCtx = execCtx
	h.executingCancel = cancel
	defer cancel()

	start := workflow.Now(ctx)
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusInProgress,
		Metadata: map[string]any{
			"execute_started_at": start.UTC().Format(time.RFC3339),
		},
	})

	hooks := h.sig.GetHooks()

	result, _ := hooks.PreExecuteHooks(ctx, signal.SignalPhaseExecute)
	if !result.Allow {
		execErr := &signal.SignalErrExecute{Err: errors.New((&signal.SignalErrBlocked{Phase: signal.SignalPhaseExecute, Reason: result.Reason}).Error())}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: execErr.Error(),
		})
		return nil, execErr
	}

	executeStart := workflow.Now(ctx)
	err := h.sig.Execute(execCtx)
	executeDur := workflow.Now(ctx).Sub(executeStart)

	hooks.PostExecuteHooks(ctx, result.Event, hooks.BuildOutcome(err, executeDur))

	if err != nil {
		if h.canceled {
			// canceled mid-execute — cancelHandler already wrote StatusCancelled
			return nil, errors.Wrap(err, "signal was canceled during execution")
		}

		execErr := &signal.SignalErrExecute{Err: err}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: execErr.Error(),
			Metadata: map[string]any{
				"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		return nil, execErr
	}

	// persist success status to DB
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusSuccess,
		Metadata: map[string]any{
			"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
		},
	})

	return nil, nil
}
