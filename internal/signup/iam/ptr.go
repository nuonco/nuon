package iam

// toPtr returns a pointer of the value, helpful for building AWS requests
func toPtr[T any](t T) *T {
	return &t
}
