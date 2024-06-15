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

func AccountIDFromContext(ctx *gin.Context) string {
	val := ctx.Value(accountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}

func FromContext(ctx *gin.Context) (*app.Account, error) {
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

func SetContext(ctx *gin.Context, acct *app.Account) {
	ctx.Set(accountCtxKey, acct)
	ctx.Set(accountIDCtxKey, acct.ID)
}
