package generics

func First[T comparable](vals ...T) T {
	empty := *new(T)
	for _, v := range vals {
		if v != empty {
			return v
		}
	}

	return empty
}
