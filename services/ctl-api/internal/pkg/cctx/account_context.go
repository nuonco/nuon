package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	accountCtxKey string = "account"
)

func AccountFromContext(ctx ValueContext) (*app.Account, error) {
	acct := ctx.Value(accountCtxKey)
	if acct == nil {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func AccountFromGinContext(ctx *gin.Context) (*app.Account, error) {
	acct, exists := ctx.Get(accountCtxKey)
	if !exists {
		return nil, fmt.Errorf("account was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func SetAccountGinContext(ctx *gin.Context, acct *app.Account) {
	ctx.Set(accountCtxKey, acct)
	ctx.Set(accountIDCtxKey, acct.ID)
}

func SetAccountContext(ctx context.Context, acct *app.Account) context.Context {
	ctx = context.WithValue(ctx, accountCtxKey, acct)
	ctx = context.WithValue(ctx, accountIDCtxKey, acct.ID)
	return ctx
}
