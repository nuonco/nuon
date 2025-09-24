package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
	handlerworkflow "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

const EnqueueUpdateName string = "enqueue"

const enqueuUpdateType = handlerTypeUpdate

type EnqueueResponse struct {
	ID         string
	WorkflowID string
}

func (w *queue) enqueueHandler(ctx workflow.Context, sig signal.Signal) (*EnqueueResponse, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	q, err := activities.AwaitGetQueueByQueueID(ctx, w.queueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	if len(w.state.QueueRefs) > q.MaxDepth {
		return nil, errors.Wrapf(err, "unable to queue item, depth of %d reached", q.MaxDepth)
	}

	qSignal, err := activities.AwaitCreateQueueSignal(ctx, sig, &activities.CreateQueueSignalRequest{
		QueueID: q.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create queue signal")
	}

	handlerworkflow.StartHandler(ctx, qSignal.Workflow.ID, handlerworkflow.HandlerRequest{
		QueueID:       q.ID,
		QueueSignalID: qSignal.ID,
	})
	l.Info("queueing signal for processing", zap.String("workflow-id", qSignal.Workflow.ID))
	w.ch.Send(ctx, QueueRef{
		WorkflowID: qSignal.Workflow.ID,
		ID:         qSignal.ID,
	})

	// and now, we queue it in our internal queueu
	return &EnqueueResponse{
		ID:         qSignal.ID,
		WorkflowID: qSignal.Workflow.ID,
	}, nil
}
