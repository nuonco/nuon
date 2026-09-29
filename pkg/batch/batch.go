package batch

import "context"

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
