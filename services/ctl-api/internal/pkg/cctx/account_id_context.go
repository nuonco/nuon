package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/workflow"
)

const (
	accountIDCtxKey string = "account_id"
)

func AccountIDFromContext(ctx ValueContext) (string, error) {
	acctID := ctx.Value(accountIDCtxKey)
	if acctID == nil {
		return "", fmt.Errorf("account was not set on middleware context")
	}

	return acctID.(string), nil
}

func SetAccountIDWorkflowContext(ctx workflow.Context, acctID string) workflow.Context {
	return workflow.WithValue(ctx, accountIDCtxKey, acctID)
}

func SetAccountIDContext(ctx context.Context, acctID string) context.Context {
	return context.WithValue(ctx, accountIDCtxKey, acctID)
}

func SetAccountIDGinContext(ctx *gin.Context, acctID string) {
	ctx.Set(accountIDCtxKey, acctID)
}
