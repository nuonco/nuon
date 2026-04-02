package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
)

// noRetryLongTimeout is used for activities that call workflow update handlers
// (validate, execute) which block until the signal finishes. These can run for
// extended periods and must not be retried by Temporal.
var noRetryLongTimeout = &workflow.ActivityOptions{
	StartToCloseTimeout: 2 * time.Hour,
	RetryPolicy: &temporal.RetryPolicy{
		MaximumAttempts: 1,
	},
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

	// skip signals that were already processed (e.g. via force-execute)
	if isTerminalStatus(queueSignal.Status.Status) {
		l.Info("queue signal already in terminal state, skipping",
			zap.String("queue-signal-id", queueSignal.ID),
			zap.String("status", string(queueSignal.Status.Status)))
		return nil
	}

	signalErr := q.processQueueSignal(ctx, l, queueSignal, queueRef)
	if signalErr != nil {
		// Persist error status so AwaitSignal callers don't block forever
		if statusErr := activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
			QueueSignalID: queueSignal.ID,
			Status:        app.StatusError,
		}); statusErr != nil {
			l.Warn("failed to update queue signal status after error",
				zap.String("queue-signal-id", queueSignal.ID),
				zap.Error(statusErr))
		}
		return signalErr
	}

	// track this finished handler and sleep old ones
	q.state.QueueRefs = append(q.state.QueueRefs, queueRef)
	q.sleepOldHandlers(ctx)

	return nil
}

func (q *queue) processQueueSignal(ctx workflow.Context, l *zap.Logger, queueSignal *app.QueueSignal, queueRef QueueRef) error {
	l.Info("making sure queue signal workflow is ready")
	if _, err := handleractivities.AwaitUpdateWorkflowReady(ctx, handleractivities.UpdateWorkflowReadyRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
	}); err != nil {
		return errors.Wrap(err, "unable to update")
	}

	l.Info("making sure queue signal workflow is valid")
	if _, err := handleractivities.AwaitUpdateWorkflowValidate(ctx, handleractivities.UpdateWorkflowValidateRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
	}, noRetryLongTimeout); err != nil {
		return errors.Wrap(err, "unable to validate")
	}

	if _, err := handleractivities.AwaitUpdateWorkflowExecute(ctx, handleractivities.UpdateWorkflowExecuteRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
	}, noRetryLongTimeout); err != nil {
		return errors.Wrap(err, "unable to execute")
	}

	return nil
}

// defaultSleepAfterN is the number of signals that must complete after a handler
// before the handler is put to sleep.
const defaultSleepAfterN = 5

// sleepOldHandlers puts handler workflows to sleep once enough newer signals
// have completed after them. This frees Temporal resources for finished handlers.
func (q *queue) sleepOldHandlers(ctx workflow.Context) {
	l, _ := log.WorkflowLogger(ctx)

	total := len(q.state.QueueRefs)
	if total <= defaultSleepAfterN {
		return
	}

	// Sleep all handlers except the most recent defaultSleepAfterN
	toSleep := q.state.QueueRefs[:total-defaultSleepAfterN]
	q.state.QueueRefs = q.state.QueueRefs[total-defaultSleepAfterN:]

	for _, ref := range toSleep {
		if _, err := handleractivities.AwaitUpdateWorkflowSleep(ctx, handleractivities.UpdateWorkflowSleepRequest{
			UpdateID:   ref.ID,
			WorkflowID: ref.WorkflowID,
		}); err != nil {
			// Log but don't fail - the handler may have already been stopped
			if l != nil {
				l.Warn("failed to sleep handler workflow",
					zap.String("workflow-id", ref.WorkflowID),
					zap.Error(err))
			}
		}
	}
}
