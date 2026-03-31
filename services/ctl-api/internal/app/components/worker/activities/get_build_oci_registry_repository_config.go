package activities

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetComponentOCIRegistryRepository struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ComponentID
func (a *Activities) GetComponentOCIRegistryRepository(ctx context.Context, req *GetComponentOCIRegistryRepository) (*configs.OCIRegistryRepository, error) {
	comp, err := a.helpers.GetComponent(ctx, req.ComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	compApp, err := a.getApp(ctx, comp.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app")
	}

	cfg := &configs.OCIRegistryRepository{
		Repository: compApp.Repository.RepositoryURI,
		Region:     compApp.Repository.Region,
	}

	switch a.cfg.CloudProvider {
	case "gcp":
		cfg.RegistryType = configs.OCIRegistryTypeGAR
		// LoginServer is the GAR host (e.g. "us-central1-docker.pkg.dev")
		if idx := strings.Index(compApp.Repository.RepositoryURI, "/"); idx != -1 {
			cfg.LoginServer = compApp.Repository.RepositoryURI[:idx]
		}
	default:
		cfg.RegistryType = configs.OCIRegistryTypeECR
		cfg.ECRAuth = &credentials.Config{
			Region:     compApp.Repository.Region,
			UseDefault: true,
		}
	}

	return cfg, nil
}

func (a *Activities) getApp(ctx context.Context, appID string) (*app.App, error) {
	var currentApp app.App
	if res := a.db.WithContext(ctx).
		Preload("Repository").
		First(&currentApp, "id = ?", appID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get app")
	}

	return &currentApp, nil
}
