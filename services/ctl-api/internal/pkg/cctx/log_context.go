package cctx

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func SetWorkflowLoggerContext(ctx workflow.Context, fields []zap.Field) workflow.Context {
	return workflow.WithValue(ctx, keys.LoggerFieldsCtxKey, fields)
}

func SetAPILoggerFields(ctx *gin.Context, fields []zap.Field) {
	ctx.Set(keys.LoggerFieldsCtxKey, fields)
}

func GetLogger(ctx ValueContext, l *zap.Logger) *zap.Logger {
	val := ctx.Value(keys.LoggerFieldsCtxKey)
	fields := []zap.Field{}
	if val != nil {
		fields = val.([]zap.Field)
	}

	traceID := TraceIDFromContext(ctx)
	orgID, _ := OrgIDFromContext(ctx)
	acctID, _ := AccountIDFromContext(ctx)

	if traceID != "" {
		fields = append(fields, zap.String("nuon_trace_id", traceID))
	}

	if orgID != "" {
		fields = append(fields, zap.String("org_id", orgID))
	}

	if acctID != "" {
		fields = append(fields, zap.String("account_id", acctID))
	}

	return l.With(fields...)
}
