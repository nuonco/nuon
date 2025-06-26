package cctx

import (
	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/patcher"
)

func PatcherFromContext(ctx ValueContext) *patcher.Patcher {
	p := ctx.Value(keys.PatcherCtxKey)
	if p == nil {
		return nil
	}

	return p.(*patcher.Patcher)
}

func SetPatcherGinCtx(ctx *gin.Context, patcher patcher.Patcher) {
	ctx.Set(keys.PatcherCtxKey, &patcher)
}
