package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
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

	cfg := &configs.OCIRegistryRepository{
		Repository: currentApp.Repository.RepositoryURI,
		Region:     currentApp.Repository.Region,
	}

	switch a.cfg.CloudProvider {
	case string(app.CloudPlatformGCP):
		cfg.RegistryType = configs.OCIRegistryTypeGAR
		if idx := strings.Index(currentApp.Repository.RepositoryURI, "/"); idx != -1 {
			cfg.LoginServer = currentApp.Repository.RepositoryURI[:idx]
		}
	case string(app.CloudPlatformAzure):
		cfg.RegistryType = configs.OCIRegistryTypeACR
		cfg.LoginServer = a.cfg.ManagementACRRegistryURL
		cfg.ACRAuth = &azurecredentials.Config{
			UseDefault: true,
		}
	default:
		cfg.RegistryType = configs.OCIRegistryTypeECR
		cfg.ECRAuth = &credentials.Config{
			Region:     currentApp.Repository.Region,
			UseDefault: true,
		}
	}

	return cfg, nil
}
