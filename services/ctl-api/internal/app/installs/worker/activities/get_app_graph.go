package activities

import (
	"context"

	"github.com/pkg/errors"
)

type GetAppGraphRequest struct {
	InstallID string `json:"install_id" validate:"required"`

	// AppConfigID orders the graph against a config the install has not been
	// moved to yet. Callers generating steps for a pending config update must
	// set it: the install still points at the old config, whose graph has no
	// vertex for a component the update introduces, so those components would
	// be dropped from the generated steps entirely.
	AppConfigID string `json:"app_config_id"`

	Reverse bool `json:"reverse"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetAppGraph(ctx context.Context, req GetAppGraphRequest) ([]string, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, appGraphConfigID(req, install.AppConfigID), false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	fn := a.appsHelpers.GetConfigDefaultComponentOrder
	if req.Reverse {
		fn = a.appsHelpers.GetConfigReverseDefaultComponentOrder
	}

	return fn(ctx, cfg)
}

func appGraphConfigID(req GetAppGraphRequest, installAppConfigID string) string {
	if req.AppConfigID != "" {
		return req.AppConfigID
	}
	return installAppConfigID
}
