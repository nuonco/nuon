package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const userTokenCtxKey string = "user_token"

func FromContext(ctx *gin.Context) (*app.UserToken, error) {
	org, exists := ctx.Get(userTokenCtxKey)
	if !exists {
		return nil, fmt.Errorf("user was not set on middleware context")
	}

	return org.(*app.UserToken), nil
}
