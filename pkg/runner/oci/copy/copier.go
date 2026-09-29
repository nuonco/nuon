package ocicopy

import (
	"context"

	"github.com/go-playground/validator/v10"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/fx"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
)

type Copier interface {
	Copy(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error)

	CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error)

	CopyFromLocalRegistry(ctx context.Context, localTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error)
}

type copier struct {
	v   *validator.Validate
	cfg *runnerconfig.Config
}

var _ Copier = (*copier)(nil)

type CopierParams struct {
	fx.In

	V   *validator.Validate
	Cfg *runnerconfig.Config
}

func New(params CopierParams) Copier {
	return &copier{
		v:   params.V,
		cfg: params.Cfg,
	}
}
