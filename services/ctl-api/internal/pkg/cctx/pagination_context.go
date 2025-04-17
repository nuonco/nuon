package cctx

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/pagination"
)

const (
	offPaginationCtxKey string = "offset_pagination"
)

func OffsetPaginationFromContext(ctx ValueContext) *pagination.PaginationQuery {
	p := ctx.Value(offPaginationCtxKey)
	if p == nil {
		return nil
	}

	return p.(*pagination.PaginationQuery)
}

func SetOffPaginationGinCtx(ctx *gin.Context, pagination pagination.PaginationQuery) {
	ctx.Set(offPaginationCtxKey, &pagination)
}
