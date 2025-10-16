package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetAppConfigRequest struct {
	ID string `json:"id" validate:"required"`
}

// @temporal-gen activity
// @by-id ID
// @start-to-close-timeout 10s
func (a *Activities) GetAppConfig(ctx context.Context, req GetAppConfigRequest) (*app.AppConfig, error) {
	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, req.ID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	return cfg, nil
}
