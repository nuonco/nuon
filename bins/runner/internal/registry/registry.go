package registry

import (
	"context"
	"fmt"
	"os"

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

	// distribution inits otel unconditionally inside NewRegistry with no config knob to disable it,
	// and autoexport defaults to an OTLP exporter at localhost:4318 that logs a connection-refused
	// error on every flush. We run no collector and never consume its spans, so turn it off.
	if os.Getenv("OTEL_TRACES_EXPORTER") == "" {
		os.Setenv("OTEL_TRACES_EXPORTER", "none")
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
