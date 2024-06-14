package authcontext

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	accountIDCtxKey string = "account_id"
	accountCtxKey   string = "account"
)

func FromContext(ctx *gin.Context) (*app.Account, error) {
	acct, exists := ctx.Get(accountCtxKey)
	if !exists {
		return nil, fmt.Errorf("account was not set on middleware context")
	}

	return acct.(*app.Account), nil
}

func SetContext(ctx *gin.Context, acct *app.Account) {
	ctx.Set(accountCtxKey, acct)
	ctx.Set(accountIDCtxKey, acct.ID)
}
