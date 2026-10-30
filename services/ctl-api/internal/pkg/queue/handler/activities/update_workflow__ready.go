package activities

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowReady(ctx context.Context, workflowID string, updateID string, queueID string) (*handler.ReadyResponse, error) {
	info := activity.GetInfo(ctx)

	var resp *handler.ReadyResponse
	err := handler.RetryUpdate(ctx, handler.ReadyHandlerName, func(err error) bool {
		return strings.HasPrefix(err.Error(), "unable to call query handler: unknown update ready.") ||
			strings.HasPrefix(err.Error(), "unable get response: unknown update ready.")
	}, func() error {
		startOp := a.handlerStartOperation(workflowID, queueID, updateID, info.TaskQueue)
		rawResp, err := a.tclient.UpdateWithStartWorkflowInNamespace(ctx,
			info.WorkflowNamespace,
			tclient.UpdateWithStartWorkflowOptions{
				UpdateOptions: tclient.UpdateWorkflowOptions{
					UpdateID:     updateID + "-ready",
					WorkflowID:   workflowID,
					UpdateName:   handler.ReadyHandlerName,
					WaitForStage: tclient.WorkflowUpdateStageCompleted,
				},
				StartWorkflowOperation: startOp,
			})
		if err != nil {
			return errors.Wrap(err, "unable to call query handler")
		}

		var result handler.ReadyResponse
		if err := rawResp.Get(ctx, &result); err != nil {
			wrapped := errors.Wrap(err, "unable get response")
			var appErr *temporal.ApplicationError
			if errors.As(err, &appErr) && appErr.Type() == "AcceptedUpdateCompletedWorkflow" {
				return temporal.NewNonRetryableApplicationError(
					appErr.Message(),
					appErr.Type(),
					wrapped,
				)
			}
			return wrapped
		}

		// Resolve the run-id so callers can pin subsequent updates to this exact run.
		run, err := startOp.Get(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to get handler workflow run")
		}
		result.RunID = run.GetRunID()
		resp = &result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
