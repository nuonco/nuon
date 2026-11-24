package ociprune

import (
	"context"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci/prune"
)

type registry interface {
	RepositoryURL() string
	ListManifests(ctx context.Context) ([]prune.Manifest, error)
	Delete(ctx context.Context, deletions []prune.Deletion) ([]string, []prune.DeleteFailure, error)
}

type (
	openRegistryFn func(ctx context.Context, cfg *configs.OCIRegistryRepository) (registry, error)
	scanClusterFn  func(ctx context.Context, ci *kube.ClusterInfo, repositoryURL string) (*prune.InUse, error)
)

func openECRRegistry(ctx context.Context, cfg *configs.OCIRegistryRepository) (registry, error) {
	return prune.NewECRRepository(ctx, cfg)
}
