package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetActionWorkflowRequest struct {
	WorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id WorkflowID
func (a *Activities) GetActionWorkflow(ctx context.Context, req *GetActionWorkflowRequest) (*app.ActionWorkflow, error) {
	return a.getActionWorkflow(ctx, req.WorkflowID)
}

func (a *Activities) getActionWorkflow(ctx context.Context, workflowID string) (*app.ActionWorkflow, error) {
	aw := app.ActionWorkflow{}
	res := a.db.WithContext(ctx).
		First(&aw, "id = ?", workflowID)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get action workflow config")
	}

	return &aw, nil
}
