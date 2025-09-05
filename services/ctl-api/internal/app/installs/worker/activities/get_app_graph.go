package activities

import (
	"context"

	"github.com/pkg/errors"
)

type GetAppGraphRequest struct {
	InstallID string `json:"install_id" validate:"required"`

	Reverse bool `json:"reverse"`
}

// @temporal-gen activity
func (a *Activities) GetAppGraph(ctx context.Context, req GetAppGraphRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	fn := a.appsHelpers.GetConfigDefaultComponentOrder
	if req.Reverse {
		fn = a.appsHelpers.GetConfigReverseDefaultComponentOrder
	}

	return fn(ctx, cfg)
}
