package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type paginationInfo struct {
	HasNext bool `json:"hasNext"`
}

type timelinePayload struct {
	Data       any            `json:"data"`
	Pagination paginationInfo `json:"pagination"`
}

func isNotFoundErr(err error) bool {
	if c, ok := err.(interface{ Code() int }); ok {
		return c.Code() == 404
	}
	return false
}

func timelineQuery(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	return limit, offset
}

// timelineFetcher adapts a paginated fetch into an sseStreamConfig.Fetch
// emitting a single timelinePayload event. A 404 becomes an empty payload
// rather than an error.
func timelineFetcher(eventName string, fetch func(ctx context.Context) (any, bool, error)) func(context.Context) (sseFetchResult, error) {
	return func(ctx context.Context) (sseFetchResult, error) {
		data, hasNext, err := fetch(ctx)
		if err != nil {
			if !isNotFoundErr(err) {
				return sseFetchResult{}, err
			}
			data, hasNext = nil, false
		}

		ev, err := marshalEvent(eventName, timelinePayload{
			Data:       data,
			Pagination: paginationInfo{HasNext: hasNext},
		})
		if err != nil {
			return sseFetchResult{}, fmt.Errorf("marshal %s: %v: %w", eventName, err, errSSESilentRetry)
		}
		return sseFetchResult{Events: []sseEvent{ev}}, nil
	}
}
