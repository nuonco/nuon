package generics

// ToPtr generically returns a reference to the value v of type T
func ToPtr[T any](v T) *T {
	return &v
}

// FromPtrString safely returns an object from a pointer
func FromPtrStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
