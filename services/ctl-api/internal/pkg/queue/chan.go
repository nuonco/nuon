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

	return nil
}
