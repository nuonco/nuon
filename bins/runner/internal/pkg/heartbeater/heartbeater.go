package heartbeater

import (
	"context"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
	"github.com/nuonco/nuon/pkg/metrics"
)

type Params struct {
	fx.In

	APIClient        nuonrunner.Client
	Cfg              *internal.Config
	L                *zap.Logger `name:"system"`
	LC               fx.Lifecycle
	Settings         *settings.Settings
	MW               metrics.Writer
	Process          string `name:"process"`
	ProcessRegistrar *process.Registrar
}

type HeartBeater struct {
	settings         *settings.Settings
	apiClient        nuonrunner.Client
	l                *zap.Logger
	processRegistrar *process.Registrar

	// internal state
	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
	startTS  time.Time
	mw       metrics.Writer
	process  string
}

func New(params Params) (*HeartBeater, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	hb := &HeartBeater{
		settings:         params.Settings,
		l:                params.L,
		wg:               conc.NewWaitGroup(),
		startTS:          time.Now(),
		apiClient:        params.APIClient,
		ctx:              ctx,
		cancelFn:         cancelFn,
		mw:               params.MW,
		process:          params.Process,
		processRegistrar: params.ProcessRegistrar,
	}

	params.LC.Append(hb.LifecycleHook())
	return hb, nil
}
