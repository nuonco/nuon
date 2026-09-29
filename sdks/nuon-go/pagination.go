package nuon

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const pageNextHeader = "X-Nuon-Page-Next"

func applyPaginationQuery(query *models.GetPaginatedQuery) (offset, limit *int64) {
	if query == nil {
		return nil, nil
	}

	l := int64(query.Limit)
	if l == 0 {
		l = 10
	}
	o := int64(query.Offset)
	return &o, &l
}

func hasNextPage(hr *responseHeaderReader) bool {
	return hr.GetHeader(pageNextHeader) == "true"
}
