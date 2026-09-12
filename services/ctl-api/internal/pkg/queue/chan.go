package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

func (q *queue) setupChannels(ctx workflow.Context) error {
	queue, err := getQueueByID(ctx, q.queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	q.ch = workflow.NewNamedBufferedChannel(ctx, "work-queue", queue.MaxDepth)
	q.maxDepth = queue.MaxDepth
	q.maxInFlight = queue.MaxInFlight
	q.sem = workflow.NewSemaphore(ctx, int64(queue.MaxInFlight))

	return nil
}
