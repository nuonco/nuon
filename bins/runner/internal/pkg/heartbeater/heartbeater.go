package heartbeater

import (
	"context"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon-runner-go"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type Params struct {
	fx.In

	APIClient nuonrunner.Client
	Cfg       *internal.Config
	L         *zap.Logger
	LC        fx.Lifecycle
	Settings  *settings.Settings
}

type HeartBeater struct {
	settings  *settings.Settings
	apiClient nuonrunner.Client
	l         *zap.Logger

	// internal state
	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
	startTS  time.Time
}

func New(params Params) (*HeartBeater, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	hb := &HeartBeater{
		settings:  params.Settings,
		l:         params.L,
		wg:        conc.NewWaitGroup(),
		startTS:   time.Now(),
		apiClient: params.APIClient,
		ctx:       ctx,
		cancelFn:  cancelFn,
	}

	params.LC.Append(hb.LifecycleHook())
	return hb, nil
}
