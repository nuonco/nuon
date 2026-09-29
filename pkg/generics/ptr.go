package generics

func ToPtr[T any](v T) *T {
	return &v
}

func FromPtrStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
