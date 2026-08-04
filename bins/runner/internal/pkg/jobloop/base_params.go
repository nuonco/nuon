package jobloop

import (
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/drain"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/telemetryexport"
	"github.com/nuonco/nuon/pkg/metrics"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type BaseParams struct {
	fx.In

	LC         fx.Lifecycle
	Shutdowner fx.Shutdowner

	Client      nuonrunner.Client
	Settings    *settings.Settings
	Cfg         *runnerconfig.Config
	ErrRecorder *errs.Recorder
	MW          metrics.Writer

	L *zap.Logger `name:"system"`

	ProcessRegistrar *process.Registrar
	TelemetryExport  *telemetryexport.Supervisor `optional:"true"`
	Drainer          *drain.Drainer
}
