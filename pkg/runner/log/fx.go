package log

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type FXParams struct {
	fx.In

	L *zap.Logger `name:"dev"`
}

func NewFXLog(params FXParams) fxevent.Logger {
	l := &fxevent.ZapLogger{Logger: params.L}
	l.UseLogLevel(zapcore.DebugLevel)
	return l
}

func AsSystemLogger(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"system"`),
	)
}

func AsDevLogger(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`name:"dev"`),
	)
}
