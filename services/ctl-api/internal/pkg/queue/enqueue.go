package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	handlerworkflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

const EnqueueUpdateName string = "enqueue"

const EnqueueSignalName string = "enqueue"

const enqueuUpdateType = handlerTypeUpdate

type EnqueueResponse struct {
	ID           string
	WorkflowID   string
	Deduplicated bool
}

// EnqueueHandlerInput is the input to the enqueue update handler.
// The QueueSignal is created in the DB by the client before sending this update,
// so the handler only needs the IDs to start the handler workflow and queue the ref.
type EnqueueHandlerInput struct {
	QueueSignalID string `json:"queue_signal_id"`
	WorkflowID    string `json:"workflow_id"`
}

// @temporal-gen-v2 update
// @id enqueue
func (w *queue) enqueueHandler(ctx workflow.Context, input EnqueueHandlerInput) (*EnqueueResponse, error) {
	if err := w.enqueue(ctx, input); err != nil {
		return nil, err
	}

	return &EnqueueResponse{
		ID:         input.QueueSignalID,
		WorkflowID: input.WorkflowID,
	}, nil
}

func (w *queue) enqueue(ctx workflow.Context, input EnqueueHandlerInput) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	if err := workflow.Await(ctx, func() bool {
		return w.ready
	}); err != nil {
		return errors.Wrap(err, "unable to await for ready")
	}

	w.lastActivityTime = workflow.Now(ctx)

	if len(w.state.QueueRefs) > w.maxDepth {
		return errors.Errorf("unable to queue item, depth of %d reached", w.maxDepth)
	}

	if w.inFlightSignals[input.QueueSignalID] {
		l.Info("signal already in-flight via requeueSignals, skipping re-dispatch",
			zap.String("queue-signal-id", input.QueueSignalID))
		return nil
	}

	handlerworkflow.StartHandler(ctx, input.WorkflowID, handlerworkflow.HandlerRequest{
		QueueID:       w.queueID,
		QueueSignalID: input.QueueSignalID,
	})

	w.inFlightSignals[input.QueueSignalID] = true
	l.Info("queueing signal for processing", zap.String("workflow-id", input.WorkflowID))
	if !w.ch.SendAsync(QueueRef{
		WorkflowID: input.WorkflowID,
		ID:         input.QueueSignalID,
	}) {
		delete(w.inFlightSignals, input.QueueSignalID)
		l.Warn("channel full, signal will be picked up on next requeue cycle",
			zap.String("signal-id", input.QueueSignalID))
	}

	return nil
}

func (w *queue) startEnqueueSignalLoop(ctx workflow.Context) {
	sigCh := workflow.GetSignalChannel(ctx, EnqueueSignalName)
	workflow.Go(ctx, func(gCtx workflow.Context) {
		l, err := log.WorkflowLogger(gCtx)
		if err != nil {
			return
		}
		for {
			var input EnqueueHandlerInput
			if !sigCh.Receive(gCtx, &input) {
				return
			}
			if err := w.enqueue(gCtx, input); err != nil {
				l.Warn("unable to enqueue signal from enqueue signal channel",
					zap.String("queue-signal-id", input.QueueSignalID),
					zap.Error(err))
			}
		}
	})
}

// Drained signals are not dispatched — requeueSignals recovers them on the next run.
func (w *queue) drainEnqueueSignalChannel(ctx workflow.Context, l *zap.Logger) int {
	sigCh := workflow.GetSignalChannel(ctx, EnqueueSignalName)
	drained := 0
	for {
		var input EnqueueHandlerInput
		if !sigCh.ReceiveAsync(&input) {
			return drained
		}
		drained++
		l.Info("draining buffered enqueue signal before run end",
			zap.String("queue-signal-id", input.QueueSignalID))
	}
}
