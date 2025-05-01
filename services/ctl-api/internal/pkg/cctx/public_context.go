package cctx

import (
	
	"github.com/gin-gonic/gin"


	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func SetPublicContext(ctx *gin.Context, val bool) {
	ctx.Set(keys.IsPublicKey, val)
}

func IsPublic(ctx *gin.Context) bool {
	isPublic, exists := ctx.Get(keys.IsPublicKey)
	if !exists {
		return false
	}

	return isPublic.(bool)
}
