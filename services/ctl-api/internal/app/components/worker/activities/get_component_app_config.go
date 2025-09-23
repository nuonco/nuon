package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentAppConfigRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentID
func (a *Activities) GetComponentAppConfig(ctx context.Context, req *GetComponentAppConfigRequest) (*app.AppConfig, error) {
	cmp, err := a.helpers.GetComponent(ctx, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	appCfg, err := a.appsHelpers.GetAppLatestConfig(ctx, cmp.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config for component: %w", err)
	}

	return appCfg, nil
}
