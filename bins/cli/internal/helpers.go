package helpers

import "context"

// BatchFetch is a generic function to handle pagination with offset/limit pattern.
// It takes a callback function that will be called with each page's limit,
// and returns both the items for that page and a boolean indicating if there are more items.
func BatchFetch[T any](
	ctx context.Context,
	pageSize int,
	maxItems int,
	fetchFn func(ctx context.Context, offset, limit int) ([]T, bool, error),
) ([]T, error) {
	var allItems []T

	offset := 0

	for {
		items, hasMore, err := fetchFn(ctx, offset, pageSize)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, items...)

		if len(allItems) >= maxItems {
			allItems = allItems[:maxItems]
			return allItems, nil
		}

		if !hasMore {
			break
		}

		offset += pageSize
	}

	return allItems, nil
}
