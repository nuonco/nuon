package cctx

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func IsGlobal(ctx *gin.Context) bool {
	isGlobal, exists := ctx.Get(keys.IsGlobalKey)
	if !exists {
		return false
	}

	return isGlobal.(bool)
}

func SetIsGlobal(ctx *gin.Context, val bool) {
	ctx.Set(keys.IsGlobalKey, val)
}
