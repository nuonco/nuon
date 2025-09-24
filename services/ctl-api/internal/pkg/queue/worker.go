package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
)

const (
	queueReceiveTimeout time.Duration = time.Second * 1
)

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

		if err := q.handleQueueSignal(ctx, obj); err != nil {
			l.Error("error handling workflow signal", zap.Error(err))
		}
	}
}
