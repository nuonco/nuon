package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
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

	byts, err := parse.ToTerraformJSON(parse.ParseConfig{
		V:           a.v,
		BackendType: config.BackendTypeS3,
		Bytes:       []byte(appCfg.Content),
		Template:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to generate terraform from config: %w", err)
	}

	return &GenerateTerraformConfigResponse{
		Byts: byts,
	}, nil
}
