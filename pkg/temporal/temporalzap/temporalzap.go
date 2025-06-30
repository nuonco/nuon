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

	var fields []zap.Field
	var other []any
	for _, v := range keyvals {
		switch x := v.(type) {
		case zap.Field:
			fields = append(fields, x)
		default:
			other = append(other, x)
		}
	}

	if len(other)%2 != 0 {
		return append(fields, zap.Error(fmt.Errorf("odd number of keyvals pairs: %v", other)))
	}

	for i := 0; i < len(other); i += 2 {
		key, ok := other[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", other[i])
		}
		fields = append(fields, zap.Any(key, other[i+1]))
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
