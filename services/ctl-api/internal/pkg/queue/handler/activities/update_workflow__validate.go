package activities

import (
	"context"

	"github.com/pkg/errors"

	"go.temporal.io/sdk/activity"
	tclient "go.temporal.io/sdk/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler"
)

type UpdateWorkflowValidateRequest struct {
	Namespace  string
	UpdateID   string
	WorkflowID string
}

// @temporal-gen activity
func (a *Activities) UpdateWorkflowValidate(ctx context.Context, req *UpdateWorkflowValidateRequest) (*handler.ValidateResponse, error) {
	info := activity.GetInfo(ctx)

	rawResp, err := a.tclient.UpdateWorkflowInNamespace(ctx,
		info.WorkflowNamespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   req.WorkflowID,
			UpdateName:   handler.ValidateUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call query handler")
	}

	var resp handler.ValidateResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
