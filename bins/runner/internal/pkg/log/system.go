package log

import (
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.opentelemetry.io/contrib/bridges/otelzap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

// the system logger is used to log all things that should not be sent to our API via OTEL
type SystemParams struct {
	fx.In

	Cfg      *internal.Config
	LP       *log.LoggerProvider `name:"system"`
	Settings *settings.Settings
}

func NewSystem(params SystemParams) *zap.Logger {
	zapCore := otelzap.NewCore(
		"system",
		otelzap.WithLoggerProvider(params.LP),
	)
	l := zap.New(zapCore)
	return l
}
