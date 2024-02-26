package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GenerateTerraformConfigRequest struct {
	AppConfigID string `validate:"required"`
}

type GenerateTerraformConfigResponse struct {
	Byts []byte
}

func (a *Activities) GenerateTerraformConfig(ctx context.Context, req GenerateTerraformConfigRequest) (*GenerateTerraformConfigResponse, error) {
	appCfg := app.AppConfig{}
	res := a.db.WithContext(ctx).
		First(&appCfg, "id = ?", req.AppConfigID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}

	byts, err := a.helpers.GenerateTerraformJSONFromConfig(ctx, []byte(appCfg.Content), appCfg.Format)
	if err != nil {
		return nil, fmt.Errorf("unable to generate terraform from config: %w", err)
	}

	return &GenerateTerraformConfigResponse{
		Byts: byts,
	}, nil
}
