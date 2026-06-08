// Package ociresolve resolves an OCI source reference to its manifest descriptor
// without copying any bytes.
//
// Used by image builds (containerimage handler) before any pull/push, to
// detect when the upstream digest matches the previous build and the build
// can be marked as a no-op. The resolver also lists tags, which build-time
// update-policy evaluation uses to semver-select a tag before resolving.
//
// Resolve returns the manifest descriptor — for multi-platform images this is
// the manifest list (image index) descriptor, which is the right thing to
// content-address against because it remains stable across all platforms.
package ociresolve

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/oci"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

// Resolver resolves a source ref to its manifest descriptor.
type Resolver interface {
	// Resolve fetches the manifest descriptor for srcTag in srcCfg. It does not
	// pull any blob bytes — only the manifest is fetched. For multi-platform
	// images this returns the image index descriptor.
	//
	// Errors are returned for 404 (tag not found), auth failures, and network
	// problems. Callers translate these into structured failures surfaced to
	// the user via sync errors.
	Resolve(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string) (*ocispec.Descriptor, error)

	// Tags lists every tag visible in srcCfg's repository. Used at build
	// time: when a component config sets an `update_policy` semver
	// constraint, the build planner lists tags then semver-selects the
	// highest match before resolving it to a digest.
	//
	// Tags are returned in registry-defined order (the order returned by
	// the underlying registry's tag-listing API). Callers must not rely on
	// any particular ordering. The returned slice may be empty when the
	// repository exists but contains no tags.
	Tags(ctx context.Context, srcCfg *configs.OCIRegistryRepository) ([]string, error)
}

type resolver struct {
	v   *validator.Validate
	cfg *internal.Config
}

var _ Resolver = (*resolver)(nil)

type ResolverParams struct {
	fx.In

	V   *validator.Validate
	Cfg *internal.Config
}

func New(params ResolverParams) Resolver {
	return &resolver{
		v:   params.V,
		cfg: params.Cfg,
	}
}

func (r *resolver) Resolve(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string) (_ *ocispec.Descriptor, retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "resolve")
	ctx = opCtx
	defer func() { end(retErr) }()

	if srcTag == "" {
		return nil, fmt.Errorf("source tag is required")
	}

	repo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source repo")
	}

	// Resolve fetches the manifest descriptor for the given tag/digest reference.
	// For multi-platform images this returns the image index (manifest list)
	// descriptor, which is what we want to content-address against — it remains
	// stable across platforms and matches what `docker pull <tag>` would resolve to.
	desc, err := repo.Resolve(ctx, srcTag)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to resolve %q", srcTag)
	}

	return &desc, nil
}

func (r *resolver) Tags(ctx context.Context, srcCfg *configs.OCIRegistryRepository) (_ []string, retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "list-tags")
	ctx = opCtx
	defer func() { end(retErr) }()

	repo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source repo")
	}

	// oras-go pages tag listings via a callback. Accumulate every page into
	// a single flat slice for the caller. Pre-allocating is not worth it —
	// most repos have a handful of tags, and the library is fine with
	// growth.
	var tags []string
	if err := repo.Tags(ctx, "", func(page []string) error {
		tags = append(tags, page...)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "unable to list tags")
	}

	return tags, nil
}
