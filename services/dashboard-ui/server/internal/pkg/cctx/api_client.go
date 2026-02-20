package cctx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx/keys"
)

func APIClientFromGinContext(ctx *gin.Context) (nuon.Client, error) {
	v, exists := ctx.Get(keys.APIClientKey)
	if !exists {
		return nil, fmt.Errorf("api_client not set on context")
	}
	return v.(nuon.Client), nil
}

func SetAPIClientGinContext(ctx *gin.Context, client nuon.Client) {
	ctx.Set(keys.APIClientKey, client)
}
