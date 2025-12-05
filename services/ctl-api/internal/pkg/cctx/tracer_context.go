package cctx

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func TraceIDFromContext(ctx ValueContext) string {
	traceID := ctx.Value(keys.TraceIDCtxKey)
	if traceID == nil {
		return ""
	}

	return traceID.(string)
}

func SetTraceIDGinContext(ctx *gin.Context, traceID string) {
	ctx.Set(keys.TraceIDCtxKey, traceID)
}

func SetTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, keys.TraceIDCtxKey, traceID)
}

func SetTraceIDWorkflowContext(ctx workflow.Context, traceID string) workflow.Context {
	return workflow.WithValue(ctx, keys.TraceIDCtxKey, traceID)
}
