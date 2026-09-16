package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

func (q *queue) setupChannels(ctx workflow.Context) error {
	queue, err := activities.AwaitGetQueueByQueueID(ctx, q.queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	q.ch = workflow.NewNamedBufferedChannel(ctx, "work-queue", queue.MaxDepth)
	q.maxDepth = queue.MaxDepth
	q.maxInFlight = queue.MaxInFlight
	q.sem = workflow.NewSemaphore(ctx, int64(queue.MaxInFlight))

	q.bufferedDispatch = workflow.GetVersion(ctx, "queue-backlog-overflow", workflow.DefaultVersion, 1) != workflow.DefaultVersion
	if q.bufferedDispatch {
		workflow.Go(ctx, func(ctx workflow.Context) {
			for {
				if err := workflow.Await(ctx, func() bool {
					return q.stopped || len(q.pendingSignals) > 0
				}); err != nil || q.stopped {
					return
				}
				ref := q.pendingSignals[0]
				q.pendingSignals[0] = QueueRef{}
				q.pendingSignals = q.pendingSignals[1:]
				q.ch.Send(ctx, ref)
			}
		})
	}

	return nil
}

// Retain overflow without blocking initialization or enqueue handlers. The DB
// remains the source of truth when pending references cross a continue-as-new.
func (q *queue) dispatchSignal(ref QueueRef) bool {
	if !q.bufferedDispatch {
		return q.ch.SendAsync(ref)
	}
	q.pendingSignals = append(q.pendingSignals, ref)
	return true
}
