package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	CancelUpdateName string = "cancel"
	CancelUpdateType        = handlerTypeUpdate
)

var ErrAlreadyExecuted = errors.New("signal already succeeded")

type CancelRequest struct{}

type CancelResponse struct{}

func (h *handler) cancelHandler(ctx workflow.Context, req *CancelRequest) (*CancelResponse, error) {
	if h.finished {
		return nil, ErrAlreadyExecuted
	}

	if l, logErr := log.WorkflowLogger(ctx); logErr == nil {
		l.Info("signal-callback for cancel", zap.String("queue-signal-id", h.queueSignalID))
	}

	start := workflow.Now(ctx)
	event := h.buildSignalPhaseEvent(signal.SignalPhaseCancel)

	h.canceled = true

	// cancel executing context if mid-execute
	if h.executingCtx != nil {
		h.executingCancel()
	}

	// call signal-level cancel callback if implemented
	cancelCallbackInvoked := false
	if sc, ok := h.sig.(signal.SignalWithCancel); ok {
		if err := sc.Cancel(ctx); err != nil {
			if l, logErr := log.WorkflowLogger(ctx); logErr == nil {
				l.Warn("signal cancel callback failed", zap.Error(err))
			}
		} else {
			cancelCallbackInvoked = true
		}
	}

	// persist cancelled status to DB
	statusReq := statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusCancelled,
	}
	if cancelCallbackInvoked {
		statusReq.StatusDescription = "cancel-callback-invoked"
	}
	_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(ctx, statusReq)

	dur := workflow.Now(ctx).Sub(start)
	h.runAfterPhaseSafe(ctx, event, signal.SignalPhaseOutcome{
		Status:   signal.SignalStatusCancelled,
		Duration: dur,
	})

	h.setFinished(app.StatusCancelled, "")
	h.stopped = true

	return &CancelResponse{}, nil
}
