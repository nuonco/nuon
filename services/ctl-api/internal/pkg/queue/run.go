package queue

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/workflowmanager"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

var statusTimestampKeys = []string{
	"stopped_at",
	"finished_at",
}

const (
	QueueStatusReady           = "ready"
	QueueStatusRestartAccepted = "restart-accepted, pending signals"
	QueueStatusRestarted       = "restarted"
	QueueStatusIdle            = "idle"
	QueueStatusStopped         = "stopped"
)

var maxAliveTime time.Duration = time.Hour * 24 * 7

func (q *queue) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	l.Info("ensuring queue is active")
	if err := q.ensureActive(ctx); err != nil {
		return false, errors.Wrap(err, "unable to ensure queue is active")
	}
	if q.stopped {
		if err := q.setStatusTimestamp(ctx, l, "stopped_at"); err != nil {
			l.Warn("unable to set stopped_at metadata", zap.Error(err))
		}
		return true, nil
	}

	// Clear any stale restart hint so this run doesn't immediately restart.
	if err := activities.AwaitClearRestartHint(ctx, activities.ClearRestartHintRequest{
		QueueID: q.queueID,
	}); err != nil {
		l.Warn("unable to clear restart hint", zap.Error(err))
	}

	l.Info("registering handlers")
	if err := q.registerHandlers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to register handlers")
	}

	l.Info("setting up queue channels")
	if err := q.setupChannels(ctx); err != nil {
		return false, errors.Wrap(err, "unable to setup channels")
	}

	l.Info("starting dispatcher")
	if err := q.startDispatcher(ctx); err != nil {
		return false, errors.Wrap(err, "unable to start dispatcher")
	}

	l.Info("requeuing any remaining signals")
	if err := q.requeueSignals(ctx); err != nil {
		return false, errors.Wrap(err, "unable to requeue signals")
	}

	l.Info("starting lifecycle manager")
	historyMax := canDefaultHistoryMax
	if q.cfg != nil && q.cfg.QueueContinueAsNewHistoryMax > 0 {
		historyMax = q.cfg.QueueContinueAsNewHistoryMax
	}
	hintPeriod := canDefaultHintPeriod
	if q.cfg != nil && q.cfg.QueueContinueAsNewHintPeriod > 0 {
		hintPeriod = q.cfg.QueueContinueAsNewHintPeriod
	}
	terminateThreshold := historyMax + canDefaultTerminateOverhead
	mgr := workflowmanager.New(
		workflowmanager.WithHistoryMax(historyMax),
		workflowmanager.WithTerminateThreshold(terminateThreshold),
		workflowmanager.WithCheckInterval(hintPeriod),
		workflowmanager.WithMetricsWriter(q.mw),
		workflowmanager.WithAliveChecker(func(gCtx workflow.Context) (bool, error) {
			_, err := activities.AwaitGetQueueByQueueID(gCtx, q.queueID)
			if err != nil {
				if generics.IsGormErrRecordNotFound(err) {
					return false, nil
				}
				return true, nil // transient error, keep going
			}
			return true, nil
		}),
		workflowmanager.WithOnTerminated(func(gCtx workflow.Context, historyLen int) {
			l.Error("queue terminated: workflow history exceeded safety threshold",
				zap.String("queue_id", q.queueID),
				zap.Int("history_length", historyLen),
				zap.Int("terminate_threshold", terminateThreshold))
			q.setStatus(gCtx, l, QueueStatusStopped)

			tags := []string{"workflow_type:Queue"}
			q.mw.Incr(gCtx, "workflow.terminated", tags...)
			q.mw.Event(gCtx, &statsd.Event{
				Title:     "Queue workflow terminated due to excessive history",
				Text:      fmt.Sprintf("Queue %s terminated at %d events (threshold: %d)", q.queueID, historyLen, terminateThreshold),
				AlertType: statsd.Error,
				Tags:      tags,
			})
		}),
		workflowmanager.WithCANHintChecker(workflowmanager.CANHintCheckerFunc{
			CheckFn: func(gCtx workflow.Context) (bool, error) {
				return activities.AwaitCheckRestartHint(gCtx, activities.CheckRestartHintRequest{
					QueueID: q.queueID,
				})
			},
			ClearFn: func(gCtx workflow.Context) error {
				return activities.AwaitClearRestartHint(gCtx, activities.ClearRestartHintRequest{
					QueueID: q.queueID,
				})
			},
		}),
	)
	mgr.Start(ctx)

	// Bridge manager state to queue fields.
	workflow.Go(ctx, func(gCtx workflow.Context) {
		_ = workflow.Await(gCtx, func() bool {
			return mgr.Stopped || mgr.Restarted || mgr.Terminated
		})
		if mgr.Terminated {
			q.stopped = true
		}
		if mgr.Stopped {
			q.stopped = true
		}
		if mgr.Restarted {
			q.setStatus(gCtx, l, QueueStatusRestartAccepted)
			q.restarted = true
		}
	})

	q.setStatus(ctx, l, QueueStatusReady)
	q.ready = true

	if _, err := workflow.AwaitWithTimeout(ctx, maxAliveTime, func() bool {
		// Wait until active workers drain before restarting or stopping.
		return (q.restarted || q.stopped || q.isIdle(ctx))
	}); err != nil {
		return false, err
	}

	// This sets a drain timeout on the queue, such that once we've decided it needs to be idle, slept, or restarted
	// how long we will wait for existing in flight signals to finish.
	maxDrainTimeout := 5 * time.Minute
	if q.cfg != nil && q.cfg.QueueDrainTimeout > 0 {
		maxDrainTimeout = q.cfg.QueueDrainTimeout
	}
	if _, err := workflow.AwaitWithTimeout(ctx, maxDrainTimeout, func() bool {
		// Wait until active workers drain before restarting or stopping.
		return q.activeWorkers == 0
	}); err != nil {
		l.Info("drain timeout exceeded, proceeding with restart",
			zap.Int("active_workers", q.activeWorkers))
	}

	if q.restarted {
		q.setStatus(ctx, l, QueueStatusRestarted)
		return false, nil
	}

	// handle regular stop
	if q.stopped {
		if err := q.setStatusTimestamp(ctx, l, "stopped_at"); err != nil {
			l.Warn("unable to set stopped_at metadata", zap.Error(err))
		}

		q.setStatus(ctx, l, QueueStatusStopped)
		return true, nil
	}

	// handle idle functionality
	if q.isIdle(ctx) && q.activeWorkers == 0 {
		l.Info("queue is idle, terminating workflow")
		q.setStatus(ctx, l, QueueStatusIdle)
		return true, nil
	}

	if err := q.setStatusTimestamp(ctx, l, "finished_at"); err != nil {
		l.Warn("unable to set finished_at metadata", zap.Error(err))
	}
	return false, nil
}

func (q *queue) setStatus(ctx workflow.Context, l *zap.Logger, status string) {
	l.Info("setting queue status", zap.String("status", status))
	if err := statusactivities.AwaitUpdateQueueStatusV2(ctx, statusactivities.UpdateQueueStatusV2Request{
		QueueID: q.queueID,
		Status:  app.Status(status),
	}); err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			l.Warn("queue not found, skipping status update", zap.String("queue-id", q.queueID))
			q.stopped = true
			return
		}
		l.Warn("unable to set queue status", zap.String("status", status), zap.Error(err))
	}
}

func (q *queue) setStatusTimestamp(ctx workflow.Context, l *zap.Logger, key string) error {
	ts := workflow.Now(ctx).UTC().Format(time.RFC3339)
	metadata := make(map[string]*string, len(statusTimestampKeys))
	for _, k := range statusTimestampKeys {
		if k == key {
			v := ts
			metadata[k] = &v
		} else {
			metadata[k] = nil
		}
	}

	l.Info("setting queue status", zap.String("status", key))
	if err := activities.AwaitUpdateQueueMetadata(ctx, activities.UpdateQueueMetadataRequest{
		QueueID:  q.queueID,
		Metadata: metadata,
	}); err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			l.Warn("queue not found, skipping metadata update", zap.String("queue-id", q.queueID))
			q.stopped = true
			return nil
		}
		return err
	}
	return nil
}
