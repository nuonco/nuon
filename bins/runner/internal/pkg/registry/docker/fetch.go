package docker

import (
	"context"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*registry.AccessInfo, error) {
	var (
		username string
		password string
	)
	if cfg.OCIAuth != nil {
		username = cfg.OCIAuth.Username
		password = cfg.OCIAuth.Password
	}

	return &registry.AccessInfo{
		Image: cfg.Repository,
		Auth: &registry.AccessInfoAuth{
			Username:      username,
			Password:      password,
			ServerAddress: cfg.LoginServer,
		},
	}, nil
}
