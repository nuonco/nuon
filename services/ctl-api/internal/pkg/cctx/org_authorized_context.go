package cctx

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// OrgAuthorized reports whether the org gate already authorized this request
// org-wide. When false, the account is a member of the org but lacks an
// org-wide permission, so a downstream resource middleware must authorize the
// specific resource before a handler may proceed (fail-closed otherwise).
func OrgAuthorized(ctx *gin.Context) bool {
	val, exists := ctx.Get(keys.OrgAuthorizedKey)
	if !exists {
		return false
	}

	authorized, ok := val.(bool)
	return ok && authorized
}

func SetOrgAuthorized(ctx *gin.Context, val bool) {
	ctx.Set(keys.OrgAuthorizedKey, val)
}
