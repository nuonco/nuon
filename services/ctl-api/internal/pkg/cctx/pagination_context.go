package cctx

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/pagination"
)

const (
	paginationCtxKey string = "pagination"
)

func PaginationFromContext(ctx context.Context) *pagination.PaginationQuery {
	p := ctx.Value(paginationCtxKey)
	if p == nil {
		return nil
	}

	return p.(*pagination.PaginationQuery)
}

func SetPaginationGinCtx(ctx *gin.Context, pagination pagination.PaginationQuery) {
	ctx.Set(paginationCtxKey, &pagination)
}
