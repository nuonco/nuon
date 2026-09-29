package oci

import (
	"context"
	"fmt"
	"strings"

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

func GetRepo(ctx context.Context, cfg *configs.OCIRegistryRepository) (*remote.Repository, error) {
	accessInfo, err := FetchAccessInfo(ctx, cfg)
	if err != nil {
		return nil, err
	}

	repoRef := dockerhub.NormalizeReference(accessInfo.RepositoryURI())
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

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
