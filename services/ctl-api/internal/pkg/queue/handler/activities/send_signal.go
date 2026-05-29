package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"
)

// CallbackInfo describes where to send a Temporal signal when an operation
// completes. Used to replace activity-based blocking waits with zero-cost
// workflow signal channel receives.
type CallbackInfo struct {
	WorkflowID string `json:"callback_workflow_id"`
	SignalName string `json:"callback_signal_name"`
	Namespace  string `json:"callback_namespace"`
}

type SendSignalRequest struct {
	Callback CallbackInfo `json:"callback"`
	Payload  any          `json:"payload"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) SendSignal(ctx context.Context, req SendSignalRequest) error {
	l := activity.GetLogger(ctx)

	err := a.tclient.SignalWorkflowInNamespace(
		ctx,
		req.Callback.Namespace,
		req.Callback.WorkflowID,
		"", // empty RunID = latest run
		req.Callback.SignalName,
		req.Payload,
	)
	if err != nil {
		// Best-effort: the target workflow may have already terminated.
		// Log and return the error so callers can decide how to handle it.
		l.Warn("failed to send completion signal",
			zap.String("target_workflow", req.Callback.WorkflowID),
			zap.String("signal_name", req.Callback.SignalName),
			zap.Error(err))
		return errors.Wrap(err, "unable to send signal")
	}

	return nil
}
