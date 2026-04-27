package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/temporal/heartbeat"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @schedule-to-close-timeout 2h
// @heartbeat-timeout 30s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowExecute(ctx context.Context, workflowID string, updateID string, queueID string, runID string) (*handler.ExecuteResponse, error) {
	return heartbeat.WithHeartbeat(ctx, 10*time.Second, func(ctx context.Context) (*handler.ExecuteResponse, error) {
		info := activity.GetInfo(ctx)

		rawResp, err := a.tclient.UpdateWorkflowInNamespace(ctx,
			info.WorkflowNamespace,
			tclient.UpdateWorkflowOptions{
				UpdateID:     updateID + "-execute",
				WorkflowID:   workflowID,
				RunID:        runID,
				UpdateName:   handler.ExecuteUpdateName,
				WaitForStage: tclient.WorkflowUpdateStageCompleted,
			})
		if err != nil {
			return nil, errors.Wrap(err, "unable to call update handler")
		}

		var resp handler.ExecuteResponse
		if err := rawResp.Get(ctx, &resp); err != nil {
			wrapped := errors.Wrap(err, "unable get response")
			var appErr *temporal.ApplicationError
			if errors.As(err, &appErr) && appErr.Type() == "AcceptedUpdateCompletedWorkflow" {
				return nil, temporal.NewNonRetryableApplicationError(
					appErr.Message(),
					appErr.Type(),
					wrapped,
				)
			}
			return nil, wrapped
		}

		return &resp, nil
	})
}
