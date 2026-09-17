package api

import "fmt"

const (
	MCPDefaultListLimit = 20
	MCPMaxListLimit     = 100
)

// MCPListToolHint is appended to paginated list-tool descriptions so agents
// surface leftover rows instead of paging until the full set is in context.
const MCPListToolHint = " Paginated: default 20 items (max 100). If has_more is true, more rows exist; pass offset=next_offset for the next page. Do not keep paging until the list is complete unless the user asked for everything."

// MCPListPage clamps a list tool's limit/offset. A zero limit uses MCPDefaultListLimit.
func MCPListPage(limit, offset int) (int, int, error) {
	if limit == 0 {
		limit = MCPDefaultListLimit
	}
	if limit < 1 || limit > MCPMaxListLimit {
		return 0, 0, fmt.Errorf("limit must be between 1 and %d", MCPMaxListLimit)
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("offset must be >= 0")
	}
	return limit, offset, nil
}

// MCPClipList trims a Limit(limit+1) query down to the page and reports overflow.
func MCPClipList[T any](rows []T, limit int) ([]T, bool) {
	if len(rows) > limit {
		return rows[:limit], true
	}
	return rows, false
}

// MCPNextOffset is 0 when the page is complete so callers can omit it from JSON.
func MCPNextOffset(offset, limit int, hasMore bool) int {
	if !hasMore {
		return 0
	}
	return offset + limit
}
