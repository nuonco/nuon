package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func AccountIDFromContext(ctx ValueContext) (string, error) {
	acctID := ctx.Value(keys.AccountIDCtxKey)
	if acctID == nil {
		return "", fmt.Errorf("account was not set on middleware context")
	}

	return acctID.(string), nil
}

func SetAccountIDWorkflowContext(ctx workflow.Context, acctID string) workflow.Context {
	return workflow.WithValue(ctx, keys.AccountIDCtxKey, acctID)
}

func SetAccountIDContext(ctx context.Context, acctID string) context.Context {
	return context.WithValue(ctx, keys.AccountIDCtxKey, acctID)
}

func SetAccountIDGinContext(ctx *gin.Context, acctID string) {
	ctx.Set(keys.AccountIDCtxKey, acctID)
}
