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

var ErrSignalNoop = errors.New("queue signal already in terminal state")

// why: queueSignalErrorStatusGuardVersion gates the status re-read before the error
// write because in-flight histories scheduled the write without that activity.
const queueSignalErrorStatusGuardVersion = "queue-signal-error-status-guard-v1"

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

	if generics.SliceContains(queueSignal.Status.Status, []app.Status{app.StatusSuccess, app.StatusError, app.StatusCancelled}) {
		l.Info("queue signal already in terminal state, skipping",
			zap.String("queue-signal-id", queueSignal.ID),
			zap.String("status", string(queueSignal.Status.Status)))
		return ErrSignalNoop
	}

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
		// why: Cancellation is a domain outcome, not a dispatch failure — the
		// cancel handler already persisted StatusCancelled; never stamp
		// error over it.
		if callback.IsCancelled(signalErr) {
			l.Info("queue signal was cancelled",
				zap.String("queue-signal-id", queueSignal.ID))
			return nil
		}

		// why: Persist error status so callers don't block forever — unless the
		// handler already finalised the signal (e.g. cancelled mid-execute):
		// the handler's status is the meaningful one and a blanket error
		// write would corrupt it.
		shouldWriteError := true
		if workflow.GetVersion(ctx, queueSignalErrorStatusGuardVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			fresh, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(ctx, queueRef.ID)
			shouldWriteError = err != nil || !generics.SliceContains(fresh.Status.Status, []app.Status{app.StatusSuccess, app.StatusError, app.StatusCancelled})
		}
		if shouldWriteError {
			if statusErr := statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID: queueSignal.ID,
				Status:        app.StatusError,
			}); statusErr != nil {
				l.Warn("failed to update queue signal status after error",
					zap.String("queue-signal-id", queueSignal.ID),
					zap.Error(statusErr))
			}
		}
		return signalErr
	}

	return nil
}

func (q *queue) processQueueSignal(ctx workflow.Context, l *zap.Logger, queueSignal *app.QueueSignal, queueRef QueueRef) error {
	l.Info("starting handler workflow")
	readyResp, err := handleractivities.AwaitUpdateWorkflowReady(ctx, handleractivities.UpdateWorkflowReadyRequest{
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
		QueueID:    queueSignal.QueueID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to start handler")
	}

	_ = activities.LocalAwaitUpdateQueueSignalRunID(ctx, &activities.UpdateQueueSignalRunIDRequest{
		QueueSignalID: queueSignal.ID,
		RunID:         readyResp.RunID,
	})

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
