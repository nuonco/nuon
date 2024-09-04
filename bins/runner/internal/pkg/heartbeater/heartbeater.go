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

type HeartBeater struct {
	settings  *settings.Settings
	apiClient nuonrunner.Client
	l         *zap.Logger

	// internal state
	ctx     context.Context
	wg      *conc.WaitGroup
	startTS time.Time
}

func New(cfg *internal.Config,
	apiClient nuonrunner.Client,
	lc fx.Lifecycle,
	l *zap.Logger,
	ctx context.Context,
	settings *settings.Settings,
) (*HeartBeater, error) {
	hb := &HeartBeater{
		settings:  settings,
		l:         l,
		ctx:       ctx,
		wg:        conc.NewWaitGroup(),
		startTS:   time.Now(),
		apiClient: apiClient,
	}

	lc.Append(hb.LifecycleHook())
	return hb, nil
}
