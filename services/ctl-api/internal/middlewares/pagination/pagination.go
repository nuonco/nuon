package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/pagination"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	paginationEnabledHeaderKey string = "X-Nuon-Pagination-Enabled"
)

type Params struct {
	fx.In

	L  *zap.Logger
	DB *gorm.DB `name:"psql"`
}

type middleware struct {
	l  *zap.Logger
	db *gorm.DB
}

func (m middleware) Name() string {
	return "pagination"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		paginationEnable := ctx.Request.Header.Get(paginationEnabledHeaderKey)

		if paginationEnable == "" && paginationEnable != "true" {
			ctx.Next()
			return
		}

		offset := 0
		limit := 10
		maxLimit := 100

		if offsetParam := ctx.Query("offset"); offsetParam != "" {
			if parsedOffset, err := strconv.Atoi(offsetParam); err == nil {
				offset = parsedOffset
			}
		}

		if limitParam := ctx.Query("limit"); limitParam != "" {
			if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
				if parsedLimit > maxLimit {
					limit = maxLimit
				} else {
					limit = parsedLimit
				}
			}
		}

		paginationQuery := pagination.PaginationQuery{
			Offset: offset,
			Limit:  limit,
		}

		cctx.SetPaginationGinCtx(ctx, paginationQuery)
		ctx.Next()
	}
}

func New(params Params) *middleware {
	return &middleware{
		l: params.L,
	}
}
