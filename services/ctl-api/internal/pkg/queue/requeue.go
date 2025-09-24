package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
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
		w.ch.Send(ctx, QueueRef{
			WorkflowID: queueSignal.Workflow.ID,
			ID:         queueSignal.ID,
		})
	}

	return nil
}
