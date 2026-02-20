package cctx

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx/keys"
)

func TraceIDFromGinContext(ctx *gin.Context) string {
	v, exists := ctx.Get(keys.TraceIDKey)
	if !exists {
		return ""
	}
	return v.(string)
}

func SetTraceIDGinContext(ctx *gin.Context, traceID string) {
	ctx.Set(keys.TraceIDKey, traceID)
}
