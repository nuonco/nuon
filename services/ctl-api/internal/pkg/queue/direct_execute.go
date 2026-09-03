package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const DirectExecuteUpdateName string = "direct-execute"

type DirectExecuteRequest struct {
	QueueSignalID string `json:"queue_signal_id"`
}

type DirectExecuteResponse struct {
	QueueSignalID string `json:"queue_signal_id"`
}

// @temporal-gen-v2 update
// @id direct-execute
func (q *queue) directExecuteHandler(ctx workflow.Context, req DirectExecuteRequest) (*DirectExecuteResponse, error) {
	if err := workflow.Await(ctx, func() bool {
		return q.ready
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	q.lastActivityTime = workflow.Now(ctx)

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("direct executing queue signal", zap.String("queue-signal-id", req.QueueSignalID))

	queueSignal, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(ctx, req.QueueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	// check that the signal belongs to this queue
	if queueSignal.QueueID != q.queueID {
		return nil, errors.New("queue signal does not belong to this queue")
	}

	// if already in a terminal state, nothing to do
	if generics.SliceContains(queueSignal.Status.Status, []app.Status{app.StatusSuccess, app.StatusError, app.StatusCancelled}) {
		return &DirectExecuteResponse{QueueSignalID: req.QueueSignalID}, nil
	}

	queueRef := QueueRef{
		WorkflowID: queueSignal.Workflow.ID,
		ID:         queueSignal.ID,
	}

	// process immediately, bypassing the channel
	signalErr := q.processQueueSignal(ctx, l, queueSignal, queueRef)
	if signalErr != nil {
		if statusErr := statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID: queueSignal.ID,
			Status:        app.StatusError,
		}); statusErr != nil {
			l.Warn("failed to update queue signal status after error",
				zap.String("queue-signal-id", queueSignal.ID),
				zap.Error(statusErr))
		}
		return nil, errors.Wrap(signalErr, "direct execute failed")
	}

	q.state.QueueRefs = append(q.state.QueueRefs, queueRef)

	return &DirectExecuteResponse{QueueSignalID: req.QueueSignalID}, nil
}
