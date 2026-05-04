package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// ErrSignalNoop is returned when a signal has already been processed (e.g. via direct-execute)
// and the dispatcher encounters it again. Callers should check for this error and skip gracefully.
var ErrSignalNoop = errors.New("queue signal already in terminal state")

// longRunningActivityOptions is used for activities that call workflow update handlers
// (validate, execute) which block until the signal finishes. These can run for
// extended periods. Heartbeating ensures Temporal can detect dead workers and retry.
// Idempotent update IDs ensure retried update calls are deduplicated by Temporal.
var longRunningActivityOptions = &workflow.ActivityOptions{
	StartToCloseTimeout:    5 * time.Minute,
	ScheduleToCloseTimeout: 2 * time.Hour,
	HeartbeatTimeout:       10 * time.Second,
}

func (q *queue) handleQueueSignal(ctx workflow.Context, queueRef QueueRef) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("starting processing of queue signal")
	queueSignal, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, queueRef.ID)
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

	if queueSignal.ExecutionCount > 0 {
		l.Info("re-executing a signal that was already executed",
			zap.Any("status", queueSignal.Status),
			zap.String("queue-signal-id", queueSignal.ID),
			zap.String("status", string(queueSignal.Status.Status)))
	}

	// Mark the signal as dequeued with timestamp
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: queueSignal.ID,
		Status:        app.StatusInProgress,
		Metadata: map[string]any{
			"dequeued_at": workflow.Now(ctx).UTC().Format(time.RFC3339),
		},
	})

	signalErr := q.processQueueSignal(ctx, l, queueSignal, queueRef)
	if signalErr != nil {
		// Persist error status so AwaitSignal callers don't block forever
		if statusErr := statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
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

func (q *queue) processQueueSignal(ctx workflow.Context, l *zap.Logger, queueSignal *app.QueueSignal, queueRef QueueRef) error {
	l.Info("making sure queue signal workflow is ready")
	readyResp, err := handleractivities.AwaitUpdateWorkflowReady(ctx, handleractivities.UpdateWorkflowReadyRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to update")
	}

	// Persist the handler's RunID so client activities (FetchSteps, AwaitSignal)
	// can recover it on retry without re-calling Ready.
	_ = activities.AwaitUpdateQueueSignalRunID(ctx, &activities.UpdateQueueSignalRunIDRequest{
		QueueSignalID: queueSignal.ID,
		RunID:         readyResp.RunID,
	})

	l.Info("making sure queue signal workflow is valid")
	if _, err := handleractivities.AwaitUpdateWorkflowValidate(ctx, handleractivities.UpdateWorkflowValidateRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
		RunID:      readyResp.RunID,
	}, longRunningActivityOptions); err != nil {
		return errors.Wrap(err, "unable to validate")
	}

	executeOpts := longRunningActivityOptions

	// Fetch the deserialized signal to check for a custom timeout.
	// The queueSignal from the DB activity has Signal excluded (temporaljson:"-"),
	// so we fetch it separately via the signal-specific activity.
	sig, sigErr := activities.AwaitGetQueueSignalSignalByQueueSignalID(ctx, queueSignal.ID)
	if sigErr == nil {
		if t, ok := sig.(signal.SignalWithTimeout); ok && t.Timeout() > 0 {
			custom := *longRunningActivityOptions
			custom.ScheduleToCloseTimeout = t.Timeout()
			executeOpts = &custom
		}
	}

	if _, err := handleractivities.AwaitUpdateWorkflowExecute(ctx, handleractivities.UpdateWorkflowExecuteRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
		RunID:      readyResp.RunID,
	}, executeOpts); err != nil {
		return errors.Wrap(err, "unable to execute")
	}

	return nil
}
