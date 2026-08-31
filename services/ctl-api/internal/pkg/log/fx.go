package log

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func NewFXLog(cfg *internal.Config) (fxevent.Logger, error) {
	newLogger := zap.NewProduction
	if cfg.LogLevel == "DEBUG" {
		newLogger = zap.NewDevelopment
	}
	zl, err := newLogger()
	if err != nil {
		return nil, err
	}

	fxzl := &fxevent.ZapLogger{
		Logger: zl,
	}
	fxzl.UseErrorLevel(zapcore.ErrorLevel)
	fxzl.UseLogLevel(zapcore.DebugLevel)

	return fxzl, nil
}
