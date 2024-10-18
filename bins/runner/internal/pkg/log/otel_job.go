package log

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/bins/runner/internal"
)

func NewOTELJobLogger(cfg *internal.Config, lp *log.LoggerProvider) (*zap.Logger, error) {
	zapCore := otelzap.NewCore("oteljob",
		otelzap.WithLoggerProvider(lp))

	dev, err := NewDev(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dev logger")
	}

	// if running inside of nuonctl, we automatically print all logs to stdout as well
	if cfg.IsNuonctl {
		double := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(dev.Core(), zapCore)
		})

		return zap.New(zapCore, double), nil
	}

	return zap.New(zapCore), nil
}
