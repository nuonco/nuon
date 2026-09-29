package settings

import (
	"context"
	"log/slog"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
)

type Settings struct {
	HeartBeatTimeout time.Duration `validate:"required"`

	JobLoopMinPollPeriod              time.Duration `validate:"required"`
	SandboxMode                       bool
	LongPollJobs                      bool
	TelemetryRelayEndpoint            string
	VendorTelemetryEnabled            bool
	VendorTelemetryResourceAttributes map[string]string

	EnableLogging bool
	LoggingLevel  slog.Level `validate:"required"`
	OtelSchemaURL string
	EnableMetrics bool
	EnableSentry  bool
	Groups        []string `validate:"required"`

	Metadata map[string]string

	OTELConfiguration string `validate:"required"`

	ContainerImageTag string
	ContainerImageURL string

	Platform string

	apiClient nuonrunner.Client
	l         *zap.Logger
	Cfg       *runnerconfig.Config
}

type Params struct {
	fx.In

	Cfg       *runnerconfig.Config
	APIClient nuonrunner.Client
	LC        fx.Lifecycle
}

func New(params Params) (*Settings, error) {
	settings := &Settings{
		apiClient: params.APIClient,
		Cfg:       params.Cfg,
	}

	// why: in order to allow the settings type to be used to configure _other_ dependencies, we must
	// initialize them here, instead of using a lifecycle hook. If this is initialized in a lifecycle hook, we can
	// not use the settings in any other dependency initializer (ie: New function), because the settings will not be
	// loaded yet.
	ctx := context.Background()
	ctx, cancelFn := context.WithTimeout(ctx, 3*time.Second)
	defer cancelFn()
	if err := settings.fetch(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to fetch settings")
	}

	return settings, nil
}
