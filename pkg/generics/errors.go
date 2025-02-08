package generics

// Error interface defines the Error() string method.
type Error interface {
	Error() string
}

func ErrsToStrings[T Error](val []T) []string {
	strs := make([]string, 0)
	for _, v := range val {
		strs = append(strs, v.Error())
	}

	return strs
}
