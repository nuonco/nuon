package docker

import (
	"context"
	"fmt"

	"github.com/distribution/reference"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*registry.AccessInfo, error) {
	ref, err := reference.ParseNormalizedNamed(cfg.Repository)
	if err != nil {
		return nil, fmt.Errorf("unable to parse image name: %w", err)
	}

	host := reference.Domain(ref)
	if host == "docker.io" {
		// The normalized name parse above will turn short names like "foo/bar"
		// into "docker.io/foo/bar" but the actual registry host for these
		// is "index.docker.io".
		host = "index.docker.io"
	}
	if cfg.LoginServer != "" {
		host = cfg.LoginServer
	}

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
			ServerAddress: host,
		},
	}, nil
}
