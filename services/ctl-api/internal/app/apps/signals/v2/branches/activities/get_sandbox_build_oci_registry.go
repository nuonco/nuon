package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetSandboxBuildOCIRegistryRequest struct {
	AppID string `json:"app_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetSandboxBuildOCIRegistry(ctx context.Context, req GetSandboxBuildOCIRegistryRequest) (*configs.OCIRegistryRepository, error) {
	var currentApp app.App
	if res := a.db.WithContext(ctx).
		Preload("Repository").
		First(&currentApp, "id = ?", req.AppID); res.Error != nil {
		return nil, fmt.Errorf("unable to get app %s: %w", req.AppID, res.Error)
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypeECR,
		Repository:   currentApp.Repository.RepositoryURI,
		Region:       currentApp.Repository.Region,
		ECRAuth: &credentials.Config{
			Region:     currentApp.Repository.Region,
			UseDefault: true,
		},
	}, nil
}
