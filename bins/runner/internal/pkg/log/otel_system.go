package log

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

// the system logger is used to log all things that should not be sent to our API via OTEL
type OTELSystemParams struct {
	fx.In

	Cfg       *internal.Config
	LP        *log.LoggerProvider `name:"otel"`
	DevLogger *zap.Logger         `name:"dev"`
}

func NewOTELSystem(params OTELSystemParams) (*zap.Logger, error) {
	zapCore := otelzap.NewCore("otelsystem",
		otelzap.WithLoggerProvider(params.LP))

	// if running inside of nuonctl, we automatically print all logs to stdout as well
	if params.Cfg.IsNuonctl {
		double := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(params.DevLogger.Core(), zapCore)
		})

		return zap.New(zapCore, double), nil
	}

	return zap.New(zapCore), nil
}
