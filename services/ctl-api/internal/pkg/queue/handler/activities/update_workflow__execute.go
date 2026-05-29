package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @wrapper-prefix HandlerInternal
// @by-field WorkflowID
func (a *Activities) updateWorkflowExecute(ctx context.Context, workflowID string, updateID string, queueID string, runID string, cb callback.Ref) error {
	info := activity.GetInfo(ctx)

	_, err := a.tclient.UpdateWorkflowInNamespace(ctx,
		info.WorkflowNamespace,
		tclient.UpdateWorkflowOptions{
			UpdateID:     updateID + "-execute",
			WorkflowID:   workflowID,
			RunID:        runID,
			UpdateName:   handler.ExecuteUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageAccepted,
			Args:         []any{cb},
		})
	if err != nil {
		return errors.Wrap(err, "unable to send execute update")
	}

	return nil
}
