package log

import (
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
)

func NewOTELJobLogger(lp log.LoggerProvider) *zap.Logger {
	zapCore := otelzap.NewCore("oteljob",
		otelzap.WithLoggerProvider(lp))
	l := zap.New(zapCore)
	return l
}
