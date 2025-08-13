package settings

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

type Settings struct {
	// configuration for job polling and management
	HeartBeatTimeout time.Duration `validate:"required"`

	// control jobs
	JobLoopMinPollPeriod time.Duration `validate:"required"`
	SandboxMode          bool

	// visibility settings
	EnableLogging bool
	LoggingLevel  slog.Level `validate:"required"`
	OtelSchemaURL string
	EnableMetrics bool
	EnableSentry  bool
	Groups        []string `validate:"required"`

	// Metadata is added to sentry, metrics and loggers
	Metadata map[string]string

	// otel configuration - not really being used yet, but will be coming from the API to enable fetching things
	// like cloudwatch metrics and more.
	OTELConfiguration string `validate:"required"`

	// container
	ContainerImageTag string
	ContainerImageURL string

	// platform
	Platform string

	apiClient nuonrunner.Client
	l         *zap.Logger
	Cfg       *internal.Config
}

type Params struct {
	fx.In

	Cfg       *internal.Config
	APIClient nuonrunner.Client
	LC        fx.Lifecycle
}

func New(params Params) (*Settings, error) {
	settings := &Settings{
		apiClient: params.APIClient,
		Cfg:       params.Cfg,
	}

	// NOTE(jm): in order to allow the settings type to be used to configure _other_ dependencies, we must
	// initialize them here, instead of using a lifecycle hook. If this is initialized in a lifecycle hook, we can
	// not use the settings in any other dependency initializer (ie: New function), because the settings will not be
	// loaded yet.
	ctx := context.Background()
	ctx, cancelFn := context.WithTimeout(ctx, time.Second)
	defer cancelFn()
	if err := settings.fetch(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to fetch settings")
	}

	return settings, nil
}
