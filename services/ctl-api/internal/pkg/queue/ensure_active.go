package queue

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

func (q *queue) ensureActive(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	exists, err := activities.AwaitQueueExistsByQueueID(ctx, q.queueID)
	if err != nil {
		return err
	}
	if !exists {
		l.Warn("queue not found, stopping workflow", zap.String("queue-id", q.queueID))
		q.stopped = true
	}

	return nil
}
