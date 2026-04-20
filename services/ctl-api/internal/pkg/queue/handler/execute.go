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

func (h *handler) executeHandler(ctx workflow.Context) (*ExecuteResponse, error) {
	defer func() {
		h.finished = true
		h.executingCtx = nil
		h.executingCancel = nil
	}()

	// Increment execution count to track how many times this signal has been executed.
	_ = activities.AwaitIncrementQueueSignalExecutionCount(ctx, &activities.IncrementQueueSignalExecutionCountRequest{
		QueueSignalID: h.queueSignalID,
	})

	if h.canceled {
		return nil, errors.New("signal was canceled")
	}

	event := h.buildSignalPhaseEvent(signal.SignalPhaseExecute)

	// run before-phase hooks (fail-open)
	decision := h.runBeforePhase(ctx, event)
	if !decision.Allow {
		blockedErr := &signal.SignalErrExecute{Err: errors.New("blocked by lifecycle hook: " + decision.Reason)}
		_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: blockedErr.Error(),
		})
		return nil, blockedErr
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

	err := h.runSignalExecute(execCtx)
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
					"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
				},
			})
			return nil, panicErr
		}

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

// runSignalExecute calls the user-provided signal Execute in a panic-safe boundary.
func (h *handler) runSignalExecute(ctx workflow.Context) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = signal.NewSignalErrPanic(r, "execute")
		}
	}()

	sig, err := h.checkSandboxMode(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to check sandbox mode")
	}

	return sig.Execute(ctx)
}
