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

	q.idleTimeout = queue.IdleTimeout

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
		// if the queue is paused, wait until it is resumed
		if q.paused {
			if err := workflow.Await(ctx, func() bool {
				return !q.paused || q.stopped || q.restarted
			}); err != nil {
				return err
			}
		}

		// check release window
		if q.releaseWindow != nil {
			now := workflow.Now(ctx)
			if !q.releaseWindow.IsOpen(now) {
				nextOpen := q.releaseWindow.NextOpenTime(now)
				sleepDuration := nextOpen.Sub(now)
				if sleepDuration > 0 {
					l.Info("queue is outside release window, sleeping", zap.Duration("duration", sleepDuration))
					// We use Await with timeout instead of Sleep so we can be woken up by stop/restart/pause
					if _, err := workflow.AwaitWithTimeout(ctx, sleepDuration, func() bool {
						return q.stopped || q.restarted || q.paused
					}); err != nil {
						return err
					}
					// loop again to check conditions
					continue
				}
			}
		}

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
