package middlewares

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	accountIDCtxKey string = "account_id"
	accountCtxKey   string = "account"
)

func AccountIDFromContext(ctx context.Context) (string, error) {
	acctID := ctx.Value(accountIDCtxKey)
	if acctID == nil {
		return "", fmt.Errorf("account was not set on middleware context")
	}

	return acctID.(string), nil
}

func AccountFromContext(ctx context.Context) (*app.Account, error) {
	acct := ctx.Value(accountCtxKey)
	if acct == nil {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func AccountIDFromGinContext(ctx *gin.Context) string {
	val := ctx.Value(accountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}

func FromGinContext(ctx *gin.Context) (*app.Account, error) {
	acct, exists := ctx.Get(accountCtxKey)
	if !exists {
		return nil, fmt.Errorf("account was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func SetAccountIDContext(ctx context.Context, acctID string) context.Context {
	return context.WithValue(ctx, accountIDCtxKey, acctID)
}

func SetAccountIDGinContext(ctx *gin.Context, acctID string) {
	ctx.Set(accountIDCtxKey, acctID)
}

func SetGinContext(ctx *gin.Context, acct *app.Account) {
	ctx.Set(accountCtxKey, acct)
	ctx.Set(accountIDCtxKey, acct.ID)
}

func SetContext(ctx context.Context, acct *app.Account) context.Context {
	ctx = context.WithValue(ctx, accountCtxKey, acct)
	ctx = context.WithValue(ctx, accountIDCtxKey, acct.ID)
	return ctx
}
