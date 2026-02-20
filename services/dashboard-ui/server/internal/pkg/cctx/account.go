package cctx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx/keys"
)

func AccountIDFromGinContext(ctx *gin.Context) (string, error) {
	v, exists := ctx.Get(keys.AccountIDKey)
	if !exists {
		return "", fmt.Errorf("account_id not set on context")
	}
	return v.(string), nil
}

func SetAccountIDGinContext(ctx *gin.Context, accountID string) {
	ctx.Set(keys.AccountIDKey, accountID)
}

func IsEmployeeFromGinContext(ctx *gin.Context) bool {
	v, exists := ctx.Get(keys.IsEmployeeKey)
	if !exists {
		return false
	}
	return v.(bool)
}

func SetIsEmployeeGinContext(ctx *gin.Context, isEmployee bool) {
	ctx.Set(keys.IsEmployeeKey, isEmployee)
}
