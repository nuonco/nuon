package cctx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx/keys"
)

func TokenFromGinContext(ctx *gin.Context) (string, error) {
	v, exists := ctx.Get(keys.TokenKey)
	if !exists {
		return "", fmt.Errorf("token not set on context")
	}
	return v.(string), nil
}

func SetTokenGinContext(ctx *gin.Context, token string) {
	ctx.Set(keys.TokenKey, token)
}
