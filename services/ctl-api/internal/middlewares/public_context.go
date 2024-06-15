package middlewares

import "github.com/gin-gonic/gin"

const (
	isPublicKey string = "is_public"
)

func SetPublicContext(ctx *gin.Context, val bool) {
	ctx.Set(isPublicKey, val)
}

func IsPublic(ctx *gin.Context) bool {
	isPublic, exists := ctx.Get(isPublicKey)
	if !exists {
		return false
	}

	return isPublic.(bool)
}
