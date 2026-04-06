package activities

import (
	"context"

	"github.com/pkg/errors"
)

type GetSandboxDependentAppGraphRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetSandboxDependentAppGraph(ctx context.Context, req GetSandboxDependentAppGraphRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	return a.appsHelpers.GetConfigSandboxDependentComponentOrder(ctx, cfg)
}
