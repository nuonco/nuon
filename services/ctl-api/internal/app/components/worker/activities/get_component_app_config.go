package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetComponentAppConfigRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ComponentID
func (a *Activities) GetComponentAppConfig(ctx context.Context, req *GetComponentAppConfigRequest) (*app.AppConfig, error) {
	cmp, err := a.helpers.GetComponent(ctx, req.ComponentID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get component")
	}

	appCfg, err := a.appsHelpers.GetLatestActiveAppConfig(ctx, cmp.AppID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get app config for component")
	}

	return appCfg, nil
}
