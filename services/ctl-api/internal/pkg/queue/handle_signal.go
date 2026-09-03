package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// ErrSignalNoop is returned when a signal has already been processed (e.g. via direct-execute)
// and the dispatcher encounters it again. Callers should check for this error and skip gracefully.
var ErrSignalNoop = errors.New("queue signal already in terminal state")

func (q *queue) handleQueueSignal(ctx workflow.Context, queueRef QueueRef) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("starting processing of queue signal")
	queueSignal, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(ctx, queueRef.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue signal")
	}

	// skip signals that were already processed (e.g. via direct-execute)
	if generics.SliceContains(queueSignal.Status.Status, []app.Status{app.StatusSuccess, app.StatusError, app.StatusCancelled}) {
		l.Info("queue signal already in terminal state, skipping",
			zap.String("queue-signal-id", queueSignal.ID),
			zap.String("status", string(queueSignal.Status.Status)))
		return ErrSignalNoop
	}

	// If the signal came from an emitter and emitter signals are globally disabled, noop it.
	if queueSignal.EmitterID != nil && q.cfg.DisableEmitterSignals {
		l.Info("emitter signals disabled globally, skipping",
			zap.String("queue-signal-id", queueSignal.ID),
			zap.String("queue-id", queueSignal.QueueID))
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     queueSignal.ID,
			Status:            app.StatusCancelled,
			StatusDescription: "emitter signals disabled",
		})
		return ErrSignalNoop
	}

	if queueSignal.ExecutionCount > 0 {
		l.Info("re-executing a signal that was already executed",
			zap.Any("status", queueSignal.Status),
			zap.String("queue-signal-id", queueSignal.ID),
			zap.String("status", string(queueSignal.Status.Status)))
	}

	// Mark the signal as dequeued with timestamp
	_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: queueSignal.ID,
		Status:        app.StatusInProgress,
		Metadata: map[string]any{
			"dequeued_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
		},
	})

	var signalErr error
	signalErr = q.processQueueSignal(ctx, l, queueSignal, queueRef)
	if signalErr != nil {
		// Persist error status so callers don't block forever
		if statusErr := statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID: queueSignal.ID,
			Status:        app.StatusError,
		}); statusErr != nil {
			l.Warn("failed to update queue signal status after error",
				zap.String("queue-signal-id", queueSignal.ID),
				zap.Error(statusErr))
		}
		return signalErr
	}

	return nil
}

// processQueueSignal drives the handler through ready → validate → execute.
// Each update waits for accepted only (no heartbeating). The queue waits for a
// per-phase callback after validate and execute to guarantee ordering.
func (q *queue) processQueueSignal(ctx workflow.Context, l *zap.Logger, queueSignal *app.QueueSignal, queueRef QueueRef) error {
	// 1. Ready: start the handler workflow via update-with-start (waits for completed — fast).
	l.Info("starting handler workflow")
	readyResp, err := handleractivities.AwaitUpdateWorkflowReady(ctx, handleractivities.UpdateWorkflowReadyRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to start handler")
	}

	// Persist the handler's RunID for recovery.
	_ = activities.LocalAwaitUpdateQueueSignalRunID(ctx, &activities.UpdateQueueSignalRunIDRequest{
		QueueSignalID: queueSignal.ID,
		RunID:         readyResp.RunID,
	})

	// 2. Validate: send update (accepted-only), wait for callback.
	validateCB := callback.New(ctx, queueSignal.ID+"-validate")
	l.Info("sending validate update")
	if err := handleractivities.AwaitUpdateWorkflowValidate(ctx, handleractivities.UpdateWorkflowValidateRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
		RunID:      readyResp.RunID,
		Cb:         validateCB,
	}); err != nil {
		return errors.Wrap(err, "unable to send validate update")
	}

	if _, err := callback.AwaitWithTimeout(ctx, validateCB, callback.QuickTimeout); err != nil {
		return errors.Wrap(err, "validate failed")
	}

	// 3. Execute: send update (accepted-only), wait for callback.
	// Derive the timeout from the signal so long-running signals get the full budget.
	executeTimeout := signal.DeriveTimeout(queueSignal.Signal.Signal)
	executeCB := callback.New(ctx, queueSignal.ID+"-execute")
	l.Info("sending execute update")
	if err := handleractivities.AwaitUpdateWorkflowExecute(ctx, handleractivities.UpdateWorkflowExecuteRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
		RunID:      readyResp.RunID,
		Cb:         executeCB,
	}); err != nil {
		return errors.Wrap(err, "unable to send execute update")
	}

	if _, err := callback.AwaitWithTimeout(ctx, executeCB, executeTimeout); err != nil {
		return errors.Wrap(err, "execute failed")
	}

	return nil
}
