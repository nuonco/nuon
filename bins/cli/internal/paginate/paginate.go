package paginate

const pageSize = 50

type FetchFunc[T any] func(offset, limit int) ([]T, bool, error)

func All[T any](fetch FetchFunc[T]) ([]T, error) {
	var (
		all     []T
		offset  int
		hasMore = true
	)
	for hasMore {
		items, more, err := fetch(offset, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		offset += pageSize
		hasMore = more
	}
	return all, nil
}
