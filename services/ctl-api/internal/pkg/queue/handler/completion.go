package handler

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (h *handler) sendCompletionCallbacks(ctx workflow.Context) {
	l, _ := log.WorkflowLogger(ctx)

	result := callback.Result{
		Status:            string(h.finishedStatus),
		StatusDescription: h.finishedErr,
	}

	if workflowID := completionCallbacksWorkflowID(h.sig); workflowID != "" {
		outcome, err := activities.LocalAwaitWorkflowCompletionOutcomeByWorkflowID(ctx, workflowID)
		if err != nil {
			l.Error("holding completion callbacks: unable to resolve workflow outcome",
				zap.String("workflow_id", workflowID),
				zap.Error(err))
			return
		}
		switch outcome.Status {
		case app.StatusFailedPendingRetry:
			return
		case app.StatusError, app.StatusCancelled:
			result = callback.Result{
				Status:            string(outcome.Status),
				StatusDescription: outcome.HumanDescription(),
			}
		}
	}

	qs, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID)
	if err == nil {
		h.callbacks = qs.Callbacks
		if qs.Callback.IsSet() {
			found := false
			for _, cb := range h.callbacks {
				if cb.WorkflowID == qs.Callback.WorkflowID && cb.SignalName == qs.Callback.SignalName {
					found = true
					break
				}
			}
			if !found {
				h.callbacks = append(h.callbacks, qs.Callback)
			}
		}
	}

	if !h.callbacks.IsSet() {
		return
	}

	for _, cb := range h.callbacks {
		callback.Send(ctx, l, cb, result)
	}
}

func completionCallbacksWorkflowID(sig signal.Signal) string {
	residentFlow, ok := sig.(signal.CompletionCallbacksWorkflow)
	if !ok {
		return ""
	}
	return residentFlow.CompletionCallbacksWorkflowID()
}

func (h *handler) hasCallbacks() bool {
	return h.callbacks.IsSet()
}
