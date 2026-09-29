package ociresolve

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/op"
)

type Resolver interface {
	Resolve(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string) (*ocispec.Descriptor, error)

	Tags(ctx context.Context, srcCfg *configs.OCIRegistryRepository) ([]string, error)
}

type resolver struct {
	v   *validator.Validate
	cfg *runnerconfig.Config
}

var _ Resolver = (*resolver)(nil)

type ResolverParams struct {
	fx.In

	V   *validator.Validate
	Cfg *runnerconfig.Config
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

	var tags []string
	if err := repo.Tags(ctx, "", func(page []string) error {
		tags = append(tags, page...)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "unable to list tags")
	}

	return tags, nil
}
