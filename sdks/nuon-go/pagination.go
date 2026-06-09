package nuon

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const pageNextHeader = "X-Nuon-Page-Next"

// applyPaginationQuery returns the offset/limit pointers for a paginated
// request. The server over-fetches by one internally and reports whether more
// pages exist via the X-Nuon-Page-Next header (see hasNextPage), so the client
// sends the caller's limit unchanged — adding one here would exceed the API's
// maximum allowed limit of 100.
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

// hasNextPage reports whether the server indicated more results are available.
func hasNextPage(hr *responseHeaderReader) bool {
	return hr.GetHeader(pageNextHeader) == "true"
}
