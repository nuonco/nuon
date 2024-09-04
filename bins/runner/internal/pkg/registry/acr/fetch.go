package acr

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
	"github.com/powertoolsdev/mono/pkg/azure/acr"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*registry.AccessInfo, error) {
	token, err := acr.GetRepositoryToken(ctx, cfg.ACRAuth, cfg.LoginServer)
	if err != nil {
		return nil, fmt.Errorf("unable to get acr token: %w", err)
	}

	return &registry.AccessInfo{
		Image: cfg.Repository,
		Auth: &registry.AccessInfoAuth{
			Username:      acr.DefaultACRUsername,
			Password:      token,
			ServerAddress: cfg.LoginServer,
		},
	}, nil
}
