package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func AccountFromContext(ctx ValueContext) (*app.Account, error) {
	acct := ctx.Value(keys.AccountCtxKey)
	if acct == nil {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func AccountFromGinContext(ctx *gin.Context) (*app.Account, error) {
	acct, exists := ctx.Get(keys.AccountCtxKey)
	if !exists {
		return nil, fmt.Errorf("account was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func SetAccountGinContext(ctx *gin.Context, acct *app.Account) {
	ctx.Set(keys.AccountCtxKey, acct)
	ctx.Set(keys.AccountIDCtxKey, acct.ID)
	ctx.Set(keys.IsEmployeeCtxKey, acct.IsEmployee)
}

func SetAccountContext(ctx context.Context, acct *app.Account) context.Context {
	ctx = context.WithValue(ctx, keys.AccountCtxKey, acct)
	ctx = context.WithValue(ctx, keys.AccountIDCtxKey, acct.ID)
	ctx = context.WithValue(ctx, keys.IsEmployeeCtxKey, acct.IsEmployee)
	return ctx
}
