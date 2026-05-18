package registry

import (
	"context"
	"fmt"

	ociregistry "github.com/distribution/distribution/v3/registry"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
)

type Params struct {
	fx.In

	LC  fx.Lifecycle
	Cfg *runnerconfig.Config
}

type Registry struct {
	cfg *runnerconfig.Config
	*ociregistry.Registry

	ctx      context.Context
	cancelFn func()

	wg *conc.WaitGroup
}

func New(params Params) (*Registry, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	reg := &Registry{
		wg:       conc.NewWaitGroup(),
		cfg:      params.Cfg,
		ctx:      ctx,
		cancelFn: cancelFn,
	}

	cfg := reg.getConfig(params.Cfg.RegistryPort)
	ociReg, err := ociregistry.NewRegistry(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create new registry: %w", err)
	}
	reg.Registry = ociReg

	params.LC.Append(reg.LifecycleHook())
	return reg, nil
}
