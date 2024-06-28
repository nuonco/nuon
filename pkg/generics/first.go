package generics

// First returns the first non-empty value in the provided slice.
//
// If the type of T is a pointer, the first non-nil value is returned, even
// if it is a zero value.
func First[T comparable](vals ...T) T {
	empty := *new(T)
	for _, v := range vals {
		if v != empty {
			return v
		}
	}

	return empty
}
