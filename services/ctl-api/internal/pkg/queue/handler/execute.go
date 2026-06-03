package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuecctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const ExecuteUpdateName string = "execute"

const executeUpdateType = handlerTypeUpdate

type ExecuteResponse struct{}

func (h *handler) executeHandler(ctx workflow.Context, cb callback.Ref) (resp *ExecuteResponse, retErr error) {
	l, _ := log.WorkflowLogger(ctx)
	h.executing = true
	defer func() {
		h.executing = false
		h.executingCtx = nil
		h.executingCancel = nil

		status := "success"
		desc := ""
		if retErr != nil {
			status = "error"
			desc = retErr.Error()
		}
		callback.Send(ctx, l, cb, callback.Result{Status: status, StatusDescription: desc})
	}()

	// Increment execution count to track how many times this signal has been executed.
	_ = activities.LocalAwaitIncrementQueueSignalExecutionCount(ctx, &activities.IncrementQueueSignalExecutionCountRequest{
		QueueSignalID: h.queueSignalID,
	})

	if h.finished {
		return nil, errors.New("handler already finished (validate may have failed)")
	}

	if h.canceled {
		h.setFinished(app.StatusCancelled, "signal was canceled")
		return nil, errors.New("signal was canceled")
	}

	event := h.buildSignalPhaseEvent(signal.SignalPhaseExecute)

	// run before-phase hooks (fail-open)
	decision := h.runBeforePhase(ctx, event)
	if !decision.Allow {
		blockedErr := &signal.SignalErrExecute{Err: errors.New("blocked by lifecycle hook: " + decision.Reason)}
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: blockedErr.Error(),
			Metadata: map[string]any{
				"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		h.setFinished(app.StatusError, blockedErr.Error())
		return nil, blockedErr
	}

	execCtx, cancel := workflow.WithCancel(ctx)

	// Restore the enqueuer's identity onto the execution context.
	if h.queueSignal != nil {
		execCtx = queuecctx.ApplyWorkflow(execCtx, h.queueSignal.SignalContext)
	}
	h.executingCtx = execCtx
	h.executingCancel = cancel
	defer cancel()

	start := workflow.Now(ctx)
	_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
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

	h.emitExecuteMetrics(event, err, dur)

	if err != nil {
		// If the signal panicked, write error status here (outside the panic boundary).
		var panicErr *signal.SignalErrPanic
		if errors.As(err, &panicErr) {
			_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: panicErr.Error(),
				Metadata: map[string]any{
					"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
				},
			})
			h.setFinished(app.StatusError, panicErr.Error())
			return nil, panicErr
		}

		if h.canceled {
			// canceled mid-execute — cancelHandler already wrote StatusCancelled
			h.setFinished(app.StatusCancelled, "signal was canceled during execution")
			return nil, errors.Wrap(err, "signal was canceled during execution")
		}

		execErr := &signal.SignalErrExecute{Err: err}
		humanDesc := signal.HumanError(err)
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: humanDesc,
			Metadata: map[string]any{
				"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
			},
		})
		h.setFinished(app.StatusError, humanDesc)
		return nil, temporal.NewNonRetryableApplicationError(
			"signal failure",
			humanDesc,
			execErr)
	}

	// persist success status to DB
	_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusSuccess,
		Metadata: map[string]any{
			"execute_finished_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
		},
	})

	h.setFinished(app.StatusSuccess, "")
	return nil, nil
}

// emitExecuteMetrics records the latency and execution count for the signal's
// Execute phase. Tags are sourced from the SignalPhaseEvent which already carries
// lifecycle context populated by buildSignalPhaseEvent.
func (h *handler) emitExecuteMetrics(event signal.SignalPhaseEvent, err error, dur time.Duration) {
	if h.mw == nil {
		return
	}

	status := "ok"
	switch {
	case err != nil && h.canceled:
		status = "cancelled"
	case err != nil:
		status = "error"
	}

	tags := metrics.ToTags(map[string]string{
		"signal_type": string(event.SignalType),
		"owner_type":  event.OwnerType,
		"status":      status,
	})

	h.mw.Timing("queue.signal.latency", dur, tags)
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
