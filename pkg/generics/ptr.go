package generics

// ToPtr generically returns a reference to the value v of type T
func ToPtr[T any](v T) *T {
	return &v
}
