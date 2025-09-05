package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallActionWorkflowRequest struct {
	ID string `validate:"required"`

	InstallID        string `validate:"required"`
	ActionWorkflowID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetInstallActionWorkflow(ctx context.Context, req GetInstallActionWorkflowRequest) (*app.InstallActionWorkflow, error) {
	if req.ID != "" {
		return a.getInstallActionWorkflowByID(ctx, req.ID)
	}

	return a.getInstallActionWorkflow(ctx, req.InstallID, req.ActionWorkflowID)
}

func (a *Activities) getInstallActionWorkflowByID(ctx context.Context, id string) (*app.InstallActionWorkflow, error) {
	installActionWorkflow := app.InstallActionWorkflow{}

	res := a.db.WithContext(ctx).
		First(&installActionWorkflow, "id = ?", id)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow")
	}

	return &installActionWorkflow, nil
}

func (a *Activities) getInstallActionWorkflow(ctx context.Context, installID, actionWorkflowID string) (*app.InstallActionWorkflow, error) {
	installActionWorkflow := app.InstallActionWorkflow{}

	res := a.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{
			InstallID:        installID,
			ActionWorkflowID: actionWorkflowID,
		}).
		First(&installActionWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow")
	}

	return &installActionWorkflow, nil
}
