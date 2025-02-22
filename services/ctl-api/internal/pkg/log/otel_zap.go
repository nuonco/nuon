package log

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func NewLogStreamLogger(logStream *app.LogStream, lp *log.LoggerProvider, sysLog *zap.Logger, allAttrs ...map[string]string) (*zap.Logger, error) {
	zapCore := otelzap.NewCore("workflow",
		otelzap.WithLoggerProvider(lp))

	double := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(sysLog.Core(), zapCore)
	})

	l := zap.New(zapCore, double)
	for _, attrs := range allAttrs {
		for k, v := range attrs {
			l = l.With(zap.String(k, v))
		}
	}

	return l, nil
}
