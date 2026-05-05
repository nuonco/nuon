package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	queueReceiveTimeout     time.Duration = time.Minute * 5
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

	//queueSignals, err := activities.AwaitGetQueueSignalsByQueueID(ctx, q.queueID)
	//if err != nil {
	//return false
	//}
	//if len(queueSignals) > 0 {
	//return false
	//}

	return workflow.Now(ctx).Sub(q.lastActivityTime) >= q.getIdleTimeout()
}

func (q *queue) startDispatcher(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	workflow.Go(ctx, func(gCtx workflow.Context) {
		if err := q.dispatcher(gCtx); err != nil {
			l.Error("error from dispatcher", zap.Error(err))
		}
	})

	return nil
}

func (q *queue) dispatcher(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	for {
		if q.stopped || q.restarted {
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

		// Wait until not paused before dispatching.
		if q.paused {
			if err := workflow.Await(ctx, func() bool {
				return !q.paused || q.stopped || q.restarted
			}); err != nil {
				return err
			}
			if q.stopped || q.restarted {
				return nil
			}
		}

		// Acquire semaphore permit — blocks if at MaxInFlight.
		if err := q.sem.Acquire(ctx, 1); err != nil {
			return err
		}

		// Increment activeWorkers in the dispatcher (before workflow.Go) to
		// prevent the main coroutine from observing activeWorkers == 0 and
		// triggering continue-as-new before the goroutine starts.
		q.activeWorkers++
		q.lastActivityTime = workflow.Now(ctx)

		ref := obj
		workflow.Go(ctx, func(gCtx workflow.Context) {
			defer func() {
				q.sem.Release(1)
				q.activeWorkers--
				q.lastActivityTime = workflow.Now(gCtx)
				delete(q.inFlightSignals, ref.ID)
			}()

			if err := q.handleQueueSignal(gCtx, ref); err != nil {
				if errors.Is(err, ErrSignalNoop) {
					l.Info("signal already processed, noop", zap.String("queue-ref-id", ref.ID))
				} else {
					l.Error("error handling workflow signal", zap.Error(err))
				}
			}
		})
	}
}
