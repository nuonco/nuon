package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const (
	queueReceiveTimeout     time.Duration = time.Minute * 1
	defaultQueueIdleTimeout time.Duration = time.Minute * 10
)

func (q *queue) getIdleTimeout() time.Duration {
	if q.idleTimeout > 0 {
		return q.idleTimeout
	}
	if q.cfg != nil && q.cfg.QueueIdleTimeout > 0 {
		return q.cfg.QueueIdleTimeout
	}
	return defaultQueueIdleTimeout
}

func (q *queue) isIdle(ctx workflow.Context) bool {
	if q.lastActivityTime.IsZero() || q.paused {
		return false
	}

	queueSignals, err := activities.AwaitGetQueueSignalsByQueueID(ctx, q.queueID)
	if err != nil {
		return false
	}
	if len(queueSignals) > 0 {
		return false
	}

	return workflow.Now(ctx).Sub(q.lastActivityTime) >= q.getIdleTimeout()
}

func (q *queue) startWorkers(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	queue, err := activities.AwaitGetQueueByQueueID(ctx, q.queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	for i := 0; i < queue.MaxInFlight; i++ {
		workflow.Go(ctx, func(gCtx workflow.Context) {
			if err := q.worker(gCtx); err != nil {
				l.Error("error from worker", zap.Error(err))
			}
		})
	}

	return nil
}

func (q *queue) worker(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get worker")
	}

	for {
		if q.stopped {
			return nil
		}
		if q.restarted {
			return nil
		}

		var obj QueueRef
		ok, more := q.ch.ReceiveWithTimeout(ctx, queueReceiveTimeout, &obj)
		if !more {
			return nil
		}
		if !ok {
			l.Debug("workflow is starved, waiting for more signals")
			continue
		}

		q.activeWorkers++
		q.lastActivityTime = workflow.Now(ctx)

		if err := q.handleQueueSignal(ctx, obj); err != nil {
			l.Error("error handling workflow signal", zap.Error(err))
		}

		q.activeWorkers--
		q.lastActivityTime = workflow.Now(ctx)
	}
}
