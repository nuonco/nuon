package settings

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon-runner-go"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

type Settings struct {
	HeartBeatTimeout             time.Duration `validate:"required"`
	JobExecutionHeartBeatTimeout time.Duration `validate:"required"`
	OTELConfiguration            string        `validate:"required"`
	JobLoopMinPollPeriod         time.Duration `validate:"required"`
	MaxConcurrentJobs            int           `validate:"required"`

	apiClient nuonrunner.Client
	l         *zap.Logger

	// internal state
	ctx context.Context
}

func New(cfg *internal.Config,
	ctx context.Context,
	apiClient nuonrunner.Client,
	lc fx.Lifecycle,
) (*Settings, error) {
	settings := &Settings{
		apiClient: apiClient,
		ctx:       ctx,
	}
	lc.Append(settings.LifecycleHook())

	return settings, nil
}
