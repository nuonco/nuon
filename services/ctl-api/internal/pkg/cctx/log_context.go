package cctx

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func SetLogWorkflowContext(ctx workflow.Context, l *zap.Logger) workflow.Context {
	return workflow.WithValue(ctx, keys.LoggerCtxKey, l)
}

func GetLoggerWorkflow(ctx ValueContext, l *zap.Logger) *zap.Logger {
	val := ctx.Value(keys.LoggerCtxKey)
	if val == nil {
		return nil
	}

	return val.(*zap.Logger)
}
