package oci

import (
	"context"
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	pkgregistry "github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry/acr"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry/docker"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry/ecr"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

func FetchAccessInfo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*pkgregistry.AccessInfo, error) {
	var (
		err        error
		accessInfo *pkgregistry.AccessInfo
	)

	switch cfg.RegistryType {
	case configs.OCIRegistryTypeACR:
		accessInfo, err = acr.FetchAccessInfo(ctx, cfg)
	case configs.OCIRegistryTypeECR:
		accessInfo, err = ecr.FetchAccessInfo(ctx, cfg)
	case configs.OCIRegistryTypePublicOCI, configs.OCIRegistryTypePrivateOCI:
		accessInfo, err = docker.FetchAccessInfo(ctx, cfg)
	default:
		return nil, fmt.Errorf("invalid registry type %s", cfg.RegistryType)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get %s access info: %w", cfg.RegistryType, err)
	}

	return accessInfo, nil
}

func GetRepo(ctx context.Context, cfg *configs.OCIRegistryRepository) (registry.Repository, error) {
	accessInfo, err := FetchAccessInfo(ctx, cfg)
	if err != nil {
		return nil, err
	}

	repo, err := remote.NewRepository(accessInfo.RepositoryURI())
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	var (
		username string
		password string
	)
	if accessInfo.Auth != nil {
		username = accessInfo.Auth.Username
		password = accessInfo.Auth.Password
	}
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(strings.TrimPrefix(accessInfo.Auth.ServerAddress, "https://"), auth.Credential{
			Username: username,
			Password: password,
		}),
	}

	return repo, nil
}
