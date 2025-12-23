package nuon

import (
	"github.com/nuonco/nuon-go/models"
)

func handlePagination[T any](items []T, offset, limit int64) ([]T, bool) {
	hasMore := false

	if limit == 0 {
		limit = 10
	}

	limit--

	if len(items) > int(limit) {
		items = items[:len(items)-1]
		hasMore = true
	}

	return items, hasMore
}

// handlePaginationQuery use to adjust query parameters for pagination.
func handlePaginationQuery(query *models.GetPaginatedQuery) *models.GetPaginatedQuery {
	if query != nil {
		// increase by one so we can determine if there are more items
		query.Limit += 1
		return query
	}

	return nil
}
