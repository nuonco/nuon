package registry

import (
	"context"
	"fmt"

	ociregistry "github.com/distribution/distribution/v3/registry"
	"github.com/powertoolsdev/mono/bins/runner/internal"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC     fx.Lifecycle
	Cfg    *internal.Config
	RunCtx context.Context
}

type Registry struct {
	cfg *internal.Config
	*ociregistry.Registry
}

func New(params Params) (*Registry, error) {
	reg := &Registry{
		cfg: params.Cfg,
	}

	cfg := reg.getConfig()
	ociReg, err := ociregistry.NewRegistry(params.RunCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create new registry: %w", err)
	}
	reg.Registry = ociReg

	params.LC.Append(reg.LifecycleHook())
	return reg, nil
}
