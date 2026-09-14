package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetAppConfigRequest struct {
	ID string `json:"id" validate:"required"`
}

// Skips the CLI-version gate: it is actionable at sync time, not on a
// provisioning path reading an already-accepted config.
//
// @temporal-gen-v2 activity
// @by-field ID
// @start-to-close-timeout 10s
func (a *Activities) GetAppConfig(ctx context.Context, req GetAppConfigRequest) (*app.AppConfig, error) {
	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, req.ID, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	return cfg, nil
}
