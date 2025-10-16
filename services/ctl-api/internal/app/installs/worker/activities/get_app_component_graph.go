package activities

import (
	"context"

	"github.com/pkg/errors"
)

type GetAppComponentGraphRequest struct {
	InstallID   string `json:"install_id" validate:"required"`
	ComponentID string `json:"component_id" validate:"required"`

	Reverse bool `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetAppComponentGraph(ctx context.Context, req GetAppComponentGraphRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	fn := a.appsHelpers.GetConfigComponentDeployOrder
	if req.Reverse {
		fn = a.appsHelpers.GetReverseConfigComponentDeployOrder
	}

	return fn(ctx, cfg, req.ComponentID)
}
