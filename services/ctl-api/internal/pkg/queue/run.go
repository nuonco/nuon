package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

var statusTimestampKeys = []string{
	"ready_at",
	"restarted_at",
	"stopped_at",
	"idled_at",
	"finished_at",
}

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

	l.Info("registering handlers")
	if err := q.registerHandlers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to register handlers")
	}

	l.Info("setting up queue channels")
	if err := q.setupChannels(ctx); err != nil {
		return false, errors.Wrap(err, "unable to setup channels")
	}

	l.Info("requeuing any remaining signals")
	if err := q.requeueSignals(ctx); err != nil {
		return false, errors.Wrap(err, "unable to requeue signals")
	}

	l.Info("starting dispatcher")
	if err := q.startDispatcher(ctx); err != nil {
		return false, errors.Wrap(err, "unable to start dispatcher")
	}

	if err := q.setStatusTimestamp(ctx, l, "ready_at"); err != nil {
		l.Warn("unable to set ready_at metadata", zap.Error(err))
	}

	q.ready = true

	if _, err := workflow.AwaitWithTimeout(ctx, maxAliveTime, func() bool {
		return (q.restarted || q.stopped || q.isIdle(ctx)) && q.activeWorkers == 0
	}); err != nil {
		return false, err
	}

	if q.restarted {
		if err := q.setStatusTimestamp(ctx, l, "restarted_at"); err != nil {
			l.Warn("unable to set restarted_at metadata", zap.Error(err))
		}
		return false, nil
	}
	if q.stopped {
		if err := q.setStatusTimestamp(ctx, l, "stopped_at"); err != nil {
			l.Warn("unable to set stopped_at metadata", zap.Error(err))
		}
		return true, nil
	}
	if q.isIdle(ctx) && q.activeWorkers == 0 {
		l.Info("queue is idle, terminating workflow")
		if err := q.setStatusTimestamp(ctx, l, "idled_at"); err != nil {
			l.Warn("unable to set idled_at metadata", zap.Error(err))
		}
		return true, nil
	}

	if err := q.setStatusTimestamp(ctx, l, "finished_at"); err != nil {
		l.Warn("unable to set finished_at metadata", zap.Error(err))
	}
	return false, nil
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
	return activities.AwaitUpdateQueueMetadata(ctx, activities.UpdateQueueMetadataRequest{
		QueueID:  q.queueID,
		Metadata: metadata,
	})
}
