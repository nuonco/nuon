package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const ForceExecuteUpdateName string = "force-execute"

type ForceExecuteRequest struct {
	QueueSignalID string `json:"queue_signal_id"`
}

type ForceExecuteResponse struct {
	QueueSignalID string `json:"queue_signal_id"`
}

// @temporal-gen-v2 update
// @id force-execute
func (q *queue) forceExecuteHandler(ctx workflow.Context, req ForceExecuteRequest) (*ForceExecuteResponse, error) {
	if err := workflow.Await(ctx, func() bool {
		return q.ready
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	l.Info("force executing queue signal", zap.String("queue-signal-id", req.QueueSignalID))

	queueSignal, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, req.QueueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	// check that the signal belongs to this queue
	if queueSignal.QueueID != q.queueID {
		return nil, errors.New("queue signal does not belong to this queue")
	}

	// if already in a terminal state, nothing to do
	if isTerminalStatus(queueSignal.Status.Status) {
		return &ForceExecuteResponse{QueueSignalID: req.QueueSignalID}, nil
	}

	queueRef := QueueRef{
		WorkflowID: queueSignal.Workflow.ID,
		ID:         queueSignal.ID,
	}

	// process immediately, bypassing the channel
	signalErr := q.processQueueSignal(ctx, l, queueSignal, queueRef)
	if signalErr != nil {
		if statusErr := activities.AwaitUpdateQueueSignalStatus(ctx, &activities.UpdateQueueSignalStatusRequest{
			QueueSignalID: queueSignal.ID,
			Status:        app.StatusError,
		}); statusErr != nil {
			l.Warn("failed to update queue signal status after error",
				zap.String("queue-signal-id", queueSignal.ID),
				zap.Error(statusErr))
		}
		return nil, errors.Wrap(signalErr, "force execute failed")
	}

	q.state.QueueRefs = append(q.state.QueueRefs, queueRef)

	return &ForceExecuteResponse{QueueSignalID: req.QueueSignalID}, nil
}

// isTerminalStatus returns true if the signal has finished processing.
func isTerminalStatus(s app.Status) bool {
	switch s {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled:
		return true
	default:
		return false
	}
}
