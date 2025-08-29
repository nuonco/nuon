package pagination

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/pagination"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	return "offset_pagination"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet {
			ctx.Next()
			return
		}

		offset := 0
		limit := 10
		maxLimit := 100
		page := 0

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

		if pageParam := ctx.Query("page"); pageParam != "" {
			if parsedPage, err := strconv.Atoi(pageParam); err == nil {
				page = parsedPage
			}
		}

		if page > 0 {
			offset = (page - 1) * limit
		}

		paginationQuery := pagination.PaginationQuery{
			Offset: offset,
			Limit:  limit,
			Page:   page,
		}

		cctx.SetOffPaginationGinCtx(ctx, paginationQuery)
		ctx.Next()
	}
}

func New(params Params) *middleware {
	return &middleware{
		l: params.L,
	}
}
