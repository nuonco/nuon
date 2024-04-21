package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetAppConfigRequest struct {
	AppConfigID string `validate:"required"`
}

func (a *Activities) GetAppConfig(ctx context.Context, req GetAppConfigRequest) (*app.AppConfig, error) {
	cfg := app.AppConfig{}
	res := a.db.WithContext(ctx).
		First(&cfg, "id = ?", req.AppConfigID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}

	return &cfg, nil
}
