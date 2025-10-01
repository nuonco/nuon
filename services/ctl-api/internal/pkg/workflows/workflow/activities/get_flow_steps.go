package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetFlowStepsRequest struct {
	FlowID string `json:"flow_id" validate:"required"`
}

// @temporal-gen activity
// @by-id FlowID
func (a *Activities) PkgWorkflowsFlowGetFlowSteps(ctx context.Context, req GetFlowStepsRequest) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep

	res := a.db.WithContext(ctx).
		Where(app.WorkflowStep{
			InstallWorkflowID: req.FlowID,
		}).
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	return steps, nil
}
