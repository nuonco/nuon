package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handlerworkflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
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
		// Ensure the handler workflow is running. This is a no-op if it is already alive,
		// but will restart it when the previous handler was terminated (e.g. after a hard
		// queue workflow termination).
		handlerworkflow.StartHandlerIfNeeded(ctx, queueSignal.Workflow.ID, handlerworkflow.HandlerRequest{
			QueueID:       w.queueID,
			QueueSignalID: queueSignal.ID,
		})
		w.ch.Send(ctx, QueueRef{
			WorkflowID: queueSignal.Workflow.ID,
			ID:         queueSignal.ID,
		})
	}

	return nil
}
