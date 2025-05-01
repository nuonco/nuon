package cctx

import (
	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/pagination"
)

func OffsetPaginationFromContext(ctx ValueContext) *pagination.PaginationQuery {
	p := ctx.Value(keys.OffPaginationCtxKey)
	if p == nil {
		return nil
	}

	return p.(*pagination.PaginationQuery)
}

func SetOffPaginationGinCtx(ctx *gin.Context, pagination pagination.PaginationQuery) {
	ctx.Set(keys.OffPaginationCtxKey, &pagination)
}
