package generics

func ToJSON(val string) []byte {
	var contents []byte
	if len(val) > 0 {
		contents = []byte(val)
	}

	return contents
}
