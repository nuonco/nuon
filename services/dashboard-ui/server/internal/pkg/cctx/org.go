package cctx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx/keys"
)

func OrgIDFromGinContext(ctx *gin.Context) (string, error) {
	v, exists := ctx.Get(keys.OrgIDKey)
	if !exists {
		return "", fmt.Errorf("org_id not set on context")
	}
	return v.(string), nil
}

func SetOrgIDGinContext(ctx *gin.Context, orgID string) {
	ctx.Set(keys.OrgIDKey, orgID)
}
