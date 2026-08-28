package oci

import (
	"context"
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/nuonco/nuon/pkg/oci/dockerhub"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgregistry "github.com/nuonco/nuon/pkg/runner/registry"
	"github.com/nuonco/nuon/pkg/runner/registry/acr"
	"github.com/nuonco/nuon/pkg/runner/registry/docker"
	"github.com/nuonco/nuon/pkg/runner/registry/ecr"
	"github.com/nuonco/nuon/pkg/runner/registry/gar"
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
	repo.PlainHTTP = repo.Reference.Registry == "localhost" || strings.HasPrefix(repo.Reference.Registry, "localhost:")

	// Always give every repository its own isolated auth client and cache.
	// Leaving repo.Client nil makes oras-go fall back to the process-global
	// auth.DefaultClient/auth.DefaultCache, which is shared across every job in
	// a long-lived worker process. That lets a token cached for a host by one
	// job leak into another job's request to the same host — e.g. an
	// authenticated pull caching a credential under "public.ecr.aws" (the
	// runner image and vendor public images share that host), which is then
	// attached to a later anonymous public pull and rejected with a 400
	// "Your Authorization Token is invalid". A per-repo cache keeps each pull
	// isolated. Credentials are attached only when we actually have them; for
	// anonymous pulls the nil Credential drives oras-go's anonymous bearer
	// token flow with a clean cache.
	authClient := &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}
	if accessInfo.Auth != nil && accessInfo.Auth.Username != "" {
		authClient.Credential = auth.StaticCredential(strings.TrimPrefix(accessInfo.Auth.ServerAddress, "https://"), auth.Credential{
			Username: accessInfo.Auth.Username,
			Password: accessInfo.Auth.Password,
		})
	}
	repo.Client = authClient

	return repo, nil
}
