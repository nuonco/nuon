package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallSandboxRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallSandbox(ctx context.Context, req GetInstallSandboxRequest) (*app.InstallSandbox, error) {
	is := app.InstallSandbox{}
	res := a.db.WithContext(ctx).
		Where("install_id = ?", req.InstallID).
		First(&is)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install sandbox")
	}

	return &is, nil
}
