package oci

import (
	"context"
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	pkgregistry "github.com/nuonco/nuon/bins/runner/internal/pkg/registry"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry/acr"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry/docker"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry/ecr"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/registry/gar"
	"github.com/nuonco/nuon/pkg/oci/dockerhub"
	"github.com/nuonco/nuon/pkg/plugins/configs"
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
	case configs.OCIRegistryTypeGAR:
		accessInfo, err = gar.FetchAccessInfo(ctx, cfg)
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

	// Normalize Docker Hub references (e.g., "nginx" -> "docker.io/library/nginx")
	repoRef := dockerhub.NormalizeReference(accessInfo.RepositoryURI())
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	// Only configure static credentials if we actually have them.
	// For anonymous pulls (empty credentials), rely on oras-go's default
	// anonymous bearer token flow which handles the 401 challenge properly.
	if accessInfo.Auth != nil && accessInfo.Auth.Username != "" {
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.DefaultCache,
			Credential: auth.StaticCredential(strings.TrimPrefix(accessInfo.Auth.ServerAddress, "https://"), auth.Credential{
				Username: accessInfo.Auth.Username,
				Password: accessInfo.Auth.Password,
			}),
		}
	}

	return repo, nil
}
