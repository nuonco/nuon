package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler"
)

type UpdateWorkflowExecuteRequest struct {
	Namespace  string
	UpdateID   string
	WorkflowID string
}

// @temporal-gen activity
func (a *Activities) UpdateWorkflowExecute(ctx context.Context, req *UpdateWorkflowExecuteRequest) (*handler.ExecuteResponse, error) {
	info := activity.GetInfo(ctx)

	rawResp, err := a.tclient.UpdateWorkflowInNamespace(ctx,
		info.WorkflowNamespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   req.WorkflowID,
			UpdateName:   handler.ExecuteUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call query handler")
	}

	var resp handler.ExecuteResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
