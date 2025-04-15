package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetAppConfigRequest struct {
	ID string `json:"id"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetAppConfig(ctx context.Context, req GetAppConfigRequest) (*app.AppConfig, error) {
	cfg, err := a.appsHelpers.GetFullAppConfig(ctx, req.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	return cfg, nil
}
