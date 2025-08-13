package monitor

import (
	"context"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon-runner-go"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

type Params struct {
	fx.In

	APIClient nuonrunner.Client
	Cfg       *internal.Config
	L         *zap.Logger `name:"system"`
	LC        fx.Lifecycle
	Settings  *settings.Settings
	MW        metrics.Writer
}

type Monitor struct {
	settings  *settings.Settings
	apiClient nuonrunner.Client
	l         *zap.Logger

	// internal state
	ctx      context.Context
	cancelFn func()
	wg       *conc.WaitGroup
	startTS  time.Time
	mw       metrics.Writer
}

func New(params Params) (*Monitor, error) {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)

	hb := &Monitor{
		settings:  params.Settings,
		l:         params.L,
		wg:        conc.NewWaitGroup(),
		startTS:   time.Now(),
		apiClient: params.APIClient,
		ctx:       ctx,
		cancelFn:  cancelFn,
		mw:        params.MW,
	}

	params.LC.Append(hb.LifecycleHook())
	return hb, nil
}
