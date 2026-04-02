package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	h.canceled = true
	h.finished = true
	h.stopped = true

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
	statusReq := &activities.UpdateQueueSignalStatusRequest{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusCancelled,
	}
	if cancelCallbackInvoked {
		statusReq.StatusDescription = "cancel-callback-invoked"
	}
	_ = activities.AwaitUpdateQueueSignalStatus(ctx, statusReq)

	return &CancelResponse{}, nil
}
