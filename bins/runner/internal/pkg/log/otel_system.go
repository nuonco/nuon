package log

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

// the system logger is used to log all things that should not be sent to our API via OTEL
type OTELSystemParams struct {
	fx.In

	Cfg      *internal.Config
	LP       *log.LoggerProvider `name:"otel"`
	Settings *settings.Settings
}

func NewOTELSystem(params SystemParams) *zap.Logger {
	zapCore := otelzap.NewCore("otelsystem",
		otelzap.WithLoggerProvider(params.LP))
	l := zap.New(zapCore)
	return l
}
