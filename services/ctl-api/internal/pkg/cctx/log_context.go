package cctx

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	loggerCtxKey string = "logger"
)

func SetLogWorkflowContext(ctx workflow.Context, l *zap.Logger) workflow.Context {
	return workflow.WithValue(ctx, loggerCtxKey, l)
}

func GetLoggerWorkflow(ctx ValueContext, l *zap.Logger) *zap.Logger {
	val := ctx.Value(loggerCtxKey)
	if val == nil {
		return nil
	}

	return val.(*zap.Logger)
}
