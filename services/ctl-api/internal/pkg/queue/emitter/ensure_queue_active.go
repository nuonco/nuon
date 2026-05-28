package emitter

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

func (e *emitterWorkflow) ensureQueueActive(ctx workflow.Context) error {
	l, _ := log.WorkflowLogger(ctx)

	exists, err := queueactivities.AwaitQueueExistsByQueueID(ctx, e.queueID)
	if err != nil {
		return err
	}
	if !exists {
		l.Warn("queue not found, stopping emitter", zap.String("queue-id", e.queueID))
		e.stopped = true
	}

	return nil
}
