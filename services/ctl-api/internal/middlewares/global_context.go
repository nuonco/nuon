package middlewares

import "github.com/gin-gonic/gin"

const (
	isGlobalKey string = "is_global"
)

func IsGlobal(ctx *gin.Context) bool {
	isGlobal, exists := ctx.Get(isGlobalKey)
	if !exists {
		return false
	}

	return isGlobal.(bool)
}

func SetIsGlobal(ctx *gin.Context, val bool) {
	ctx.Set(isGlobalKey, val)
}
