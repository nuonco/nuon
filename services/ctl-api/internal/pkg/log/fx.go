package log

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewFXLog() (fxevent.Logger, error) {
	zl, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	fxzl := &fxevent.ZapLogger{
		Logger: zl,
	}
	fxzl.UseErrorLevel(zapcore.DebugLevel)
	fxzl.UseLogLevel(zapcore.DebugLevel)

	return fxzl, nil
}
