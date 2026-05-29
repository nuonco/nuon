package callback

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// sendSignalRequest mirrors handler/activities.SendSignalRequest to avoid
// an import cycle. The activity is registered as "SendSignal".
type sendSignalRequest struct {
	Callback sendCallbackInfo `json:"callback"`
	Payload  any              `json:"payload"`
}

type sendCallbackInfo struct {
	WorkflowID string `json:"callback_workflow_id"`
	SignalName string `json:"callback_signal_name"`
	Namespace  string `json:"callback_namespace"`
}

// Send signals the target workflow described by ref with the given result.
// Best-effort: if the target workflow has already terminated, the error is
// logged but not propagated.
func Send(ctx workflow.Context, l *zap.Logger, ref Ref, result Result) {
	if !ref.IsSet() {
		return
	}

	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	actCtx := workflow.WithActivityOptions(ctx, actOpts)

	req := sendSignalRequest{
		Callback: sendCallbackInfo{
			WorkflowID: ref.WorkflowID,
			SignalName: ref.SignalName,
			Namespace:  ref.Namespace,
		},
		Payload: result,
	}

	err := workflow.ExecuteActivity(actCtx, "SendSignal", req).Get(ctx, nil)
	if err != nil && l != nil {
		l.Warn("completion callback failed",
			zap.String("target_workflow", ref.WorkflowID),
			zap.Error(err))
	}
}
