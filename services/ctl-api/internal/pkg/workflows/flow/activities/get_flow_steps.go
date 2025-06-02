package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetFlowStepsRequest struct {
	FlowID string `json:"flow_id"`
}

// @temporal-gen activity
// @by-id FlowID
func (a *Activities) GetFlowSteps(ctx context.Context, req GetFlowStepsRequest) ([]app.FlowStep, error) {
	// var steps []app.FlowStep
	var steps []app.InstallWorkflowStep

	res := a.db.WithContext(ctx).
		Where(app.InstallWorkflowStep{
			InstallWorkflowID: req.FlowID,
		}).
		Order("idx asc").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	return steps, nil
}
