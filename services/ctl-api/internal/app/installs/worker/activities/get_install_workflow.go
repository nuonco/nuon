package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallWorkflowRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetInstallWorkflow(ctx context.Context, req GetInstallWorkflowRequest) (*app.InstallWorkflow, error) {
	iw := app.InstallWorkflow{
		ID: req.ID,
	}
	if res := a.db.WithContext(ctx).
		First(&iw, "id = ?", req.ID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install workflow")
	}

	return &iw, nil
}
