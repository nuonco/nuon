package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SaveTerraformConfigRequest struct {
	AppConfigID   string `validate:"required"`
	TerraformJSON []byte `validate:"required"`
}

func (a *Activities) SaveTerraformConfig(ctx context.Context, req SaveTerraformConfigRequest) error {
	appCfg := app.AppConfig{
		ID: req.AppConfigID,
	}

	res := a.db.WithContext(ctx).
		Model(&appCfg).
		Updates(app.AppConfig{
			GeneratedTerraform: string(req.TerraformJSON),
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update app config: %w", res.Error)
	}

	return nil
}
