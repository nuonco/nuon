package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

func (w *queue) requeueSignals(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	// fetching jobs from the queue in the DB
	l.Info("fetching previous signals from database and requeueing them")
	queueSignals, err := activities.AwaitGetQueueSignalsByQueueID(ctx, w.queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue signals")
	}
	for _, queueSignal := range queueSignals {
		l.Info("requeuing signal", zap.String("queue-signal-id", queueSignal.ID), zap.Any("type", queueSignal.Type))
		w.ch.Send(ctx, QueueRef{
			WorkflowID: queueSignal.Workflow.ID,
			ID:         queueSignal.ID,
		})
	}

	return nil
}
