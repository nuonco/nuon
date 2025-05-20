package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentOCIRegistryRepository struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ComponentID
func (a *Activities) GetComponentOCIRegistryRepository(ctx context.Context, req *GetComponentOCIRegistryRepository) (*configs.OCIRegistryRepository, error) {
	comp, err := a.helpers.GetComponent(ctx, req.ComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	compApp, err := a.getApp(ctx, comp.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app")
	}

	return &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypeECR,
		Repository:   compApp.Repository.RepositoryURI,
		Region:       compApp.Repository.Region,
		ECRAuth: &credentials.Config{
			Region: compApp.Repository.Region,
                        UseDefault: true,
		},
	}, nil
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
