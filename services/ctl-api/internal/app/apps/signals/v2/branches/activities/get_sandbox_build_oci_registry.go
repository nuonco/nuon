package activities

import (
	"context"
	"fmt"
	"strings"

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

	cfg := &configs.OCIRegistryRepository{
		Repository: currentApp.Repository.RepositoryURI,
		Region:     currentApp.Repository.Region,
	}

	switch a.cfg.CloudProvider {
	case "gcp":
		cfg.RegistryType = configs.OCIRegistryTypeGAR
		if idx := strings.Index(currentApp.Repository.RepositoryURI, "/"); idx != -1 {
			cfg.LoginServer = currentApp.Repository.RepositoryURI[:idx]
		}
	default:
		cfg.RegistryType = configs.OCIRegistryTypeECR
		cfg.ECRAuth = &credentials.Config{
			Region: currentApp.Repository.Region,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:     a.cfg.ManagementIAMRoleARN,
				SessionName: "ctl-api-ecr-auth",
			},
		}
	}

	return cfg, nil
}
