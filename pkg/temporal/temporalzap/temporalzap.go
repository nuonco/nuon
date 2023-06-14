package temporalzap

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

type Logger struct {
	zl *zap.Logger
}

var _ log.Logger = (*Logger)(nil)

func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{
		// Skip one call frame to exclude zap_adapter itself.
		// Or it can be configured when logger is created (not always possible).
		zl: zapLogger.WithOptions(zap.AddCallerSkip(1)),
	}
}

func (log *Logger) fields(keyvals []interface{}) []zap.Field {
	if len(keyvals)%2 != 0 {
		return []zap.Field{zap.Error(fmt.Errorf("odd number of keyvals pairs: %v", keyvals))}
	}

	var fields []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	return fields
}

func (log *Logger) Debug(msg string, keyvals ...interface{}) {
	log.zl.Debug(msg, log.fields(keyvals)...)
}

func (log *Logger) Info(msg string, keyvals ...interface{}) {
	log.zl.Info(msg, log.fields(keyvals)...)
}

func (log *Logger) Warn(msg string, keyvals ...interface{}) {
	log.zl.Warn(msg, log.fields(keyvals)...)
}

func (log *Logger) Error(msg string, keyvals ...interface{}) {
	log.zl.Error(msg, log.fields(keyvals)...)
}

func (log *Logger) With(keyvals ...interface{}) *Logger {
	return &Logger{zl: log.zl.With(log.fields(keyvals)...)}
}

func (log *Logger) ZapLogger() *zap.Logger {
	return log.zl
}
