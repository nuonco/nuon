package activities

import (
	"context"
	stderrors "errors"

	"github.com/pkg/errors"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
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
// @max-retries 10
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
		l.Warn("failed to send completion signal",
			zap.String("target_workflow", req.Callback.WorkflowID),
			zap.String("signal_name", req.Callback.SignalName),
			zap.Error(err))

		var notFoundErr *serviceerror.NotFound
		if stderrors.As(err, &notFoundErr) {
			return temporal.NewNonRetryableApplicationError(
				"target workflow not found",
				"WORKFLOW_NOT_FOUND",
				err,
			)
		}

		return errors.Wrap(err, "unable to send signal")
	}

	return nil
}
