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

	start := workflow.Now(ctx)

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
	statusReq := statusactivities.UpdateQueueSignalStatusV2Request{
		QueueSignalID: h.queueSignalID,
		Status:        app.StatusCancelled,
	}
	if cancelCallbackInvoked {
		statusReq.StatusDescription = "cancel-callback-invoked"
	}
	_ = statusactivities.AwaitUpdateQueueSignalStatusV2(ctx, statusReq)

	dur := workflow.Now(ctx).Sub(start)
	hooks := h.sig.GetHooks()
	outcome := hooks.BuildOutcome(nil, dur)
	outcome.Status = signal.SignalStatusCancelled
	hooks.PostExecuteHooks(ctx, signal.SignalPhaseEvent{
		QueueSignalID: hooks.QueueSignalID,
		QueueID:       hooks.QueueID,
		SignalType:    hooks.SignalType,
		OrgID:         hooks.OrgID,
		Phase:         signal.SignalPhaseCancel,
		InstallID:     hooks.InstallID,
		ComponentID:   hooks.ComponentID,
		Operation:     hooks.Operation,
	}, outcome)

	return &CancelResponse{}, nil
}
