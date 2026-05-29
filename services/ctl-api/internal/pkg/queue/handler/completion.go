package handler

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// callbackInfo describes where to send a Temporal signal on completion.
type callbackInfo struct {
	WorkflowID string `json:"callback_workflow_id"`
	SignalName string `json:"callback_signal_name"`
	Namespace  string `json:"callback_namespace"`
}

// sendSignalRequest is the request for the SendSignal activity.
// Duplicated here to avoid an import cycle with handler/activities.
type sendSignalRequest struct {
	Callback callbackInfo `json:"callback"`
	Payload  any          `json:"payload"`
}

// sendCompletionCallbacks sends Temporal signals to the queue workflow and
// parent workflow (if configured) with the handler's terminal status. This
// replaces the blocking activity-based await pattern with zero-cost signal
// channel receives on the caller side.
//
// All sends are best-effort: if a target workflow has already terminated,
// the error is logged but not propagated. The signal's DB status is always
// persisted before this is called, so any reader will see the correct state.
func (h *handler) sendCompletionCallbacks(ctx workflow.Context) {
	l, _ := log.WorkflowLogger(ctx)

	resp := FinishedResponse{
		Status:            h.finishedStatus,
		StatusDescription: h.finishedErr,
	}

	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	actCtx := workflow.WithActivityOptions(ctx, actOpts)

	// Signal the queue workflow.
	if h.queueCallbackWorkflowID != "" {
		req := sendSignalRequest{
			Callback: callbackInfo{
				WorkflowID: h.queueCallbackWorkflowID,
				SignalName: h.queueCallbackSignalName,
				Namespace:  h.queueCallbackNamespace,
			},
			Payload: resp,
		}
		err := workflow.ExecuteActivity(actCtx, "SendSignal", req).Get(ctx, nil)
		if err != nil && l != nil {
			l.Warn("queue completion callback failed",
				zap.String("queue_signal_id", h.queueSignalID),
				zap.String("target_workflow", h.queueCallbackWorkflowID),
				zap.Error(err))
		}
	}

	// Signal the parent workflow.
	if h.parentCallbackWorkflowID != "" {
		req := sendSignalRequest{
			Callback: callbackInfo{
				WorkflowID: h.parentCallbackWorkflowID,
				SignalName: h.parentCallbackSignalName,
				Namespace:  h.parentCallbackNamespace,
			},
			Payload: resp,
		}
		err := workflow.ExecuteActivity(actCtx, "SendSignal", req).Get(ctx, nil)
		if err != nil && l != nil {
			l.Warn("parent completion callback failed",
				zap.String("queue_signal_id", h.queueSignalID),
				zap.String("target_workflow", h.parentCallbackWorkflowID),
				zap.Error(err))
		}
	}
}

// hasCallbacks returns true if any completion callbacks are configured.
func (h *handler) hasCallbacks() bool {
	return h.queueCallbackWorkflowID != "" || h.parentCallbackWorkflowID != ""
}
