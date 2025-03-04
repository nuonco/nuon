package db

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	nextHeader   = "X-Nuon-Page-Next"
	offsetHeader = "X-Nuon-Page-Offset"
	limitHeader  = "X-Nuon-Page-Limit"
)

func HandlePaginatedResponse[T any](ctx *gin.Context, models []T) ([]T, error) {
	pagination := cctx.PaginationFromContext(ctx)

	if pagination == nil {
		return models, nil
	}

	if len(models) > pagination.Limit {
		models = models[:len(models) - 1]
		ctx.Header(nextHeader, "true")
	} else {
		ctx.Header(nextHeader, "false")
	}

	ctx.Header(offsetHeader, strconv.Itoa(pagination.Offset))
	ctx.Header(limitHeader, strconv.Itoa(pagination.Limit))

	return models, nil
}
