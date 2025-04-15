package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallWorkflowStepsRequest struct {
	InstallWorkflowID string `json:"install_workflow_id"`
}

// @temporal-gen activity
// @by-id InstallWorkflowID
func (a *Activities) GetInstallWorkflowsSteps(ctx context.Context, req GetInstallWorkflowStepsRequest) ([]app.InstallWorkflowStep, error) {
	var steps []app.InstallWorkflowStep

	res := a.db.WithContext(ctx).
		Where(app.InstallWorkflowStep{
			InstallWorkflowID: req.InstallWorkflowID,
		}).
		Order("idx asc").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	return steps, nil
}
