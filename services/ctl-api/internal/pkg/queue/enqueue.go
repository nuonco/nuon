package queue

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handlerworkflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const EnqueueUpdateName string = "enqueue"

const enqueuUpdateType = handlerTypeUpdate

type EnqueueResponse struct {
	ID         string
	WorkflowID string
}

// EnqueueHandlerInput is the input to the enqueue update handler.
// OwnerID and OwnerType are optional — when set, the created QueueSignal will have
// its polymorphic owner association populated automatically (e.g. pointing at a ComponentBuild).
type EnqueueHandlerInput struct {
	Signal    signal.Signal `json:"signal"`
	OwnerID   string        `json:"owner_id,omitempty"`
	OwnerType string        `json:"owner_type,omitempty"`
}

// @temporal-gen-v2 update
// @id enqueue
func (w *queue) enqueueHandler(ctx workflow.Context, input EnqueueHandlerInput) (*EnqueueResponse, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	if err := workflow.Await(ctx, func() bool {
		return w.ready
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	w.lastActivityTime = workflow.Now(ctx)

	if len(w.state.QueueRefs) > w.maxDepth {
		return nil, errors.Errorf("unable to queue item, depth of %d reached", w.maxDepth)
	}

	qSignal, err := activities.AwaitCreateQueueSignal(ctx, &activities.CreateQueueSignalRequest{
		QueueID:   w.queueID,
		Signal:    input.Signal,
		OwnerID:   input.OwnerID,
		OwnerType: input.OwnerType,
	}, &workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create queue signal")
	}

	handlerworkflow.StartHandler(ctx, qSignal.Workflow.ID, handlerworkflow.HandlerRequest{
		QueueID:       w.queueID,
		QueueSignalID: qSignal.ID,
	})

	l.Info("queueing signal for processing", zap.String("workflow-id", qSignal.Workflow.ID))
	if !w.ch.SendAsync(QueueRef{
		WorkflowID: qSignal.Workflow.ID,
		ID:         qSignal.ID,
	}) {
		l.Warn("channel full, signal will be picked up on next requeue cycle",
			zap.String("signal-id", qSignal.ID))
	}

	return &EnqueueResponse{
		ID:         qSignal.ID,
		WorkflowID: qSignal.Workflow.ID,
	}, nil
}
