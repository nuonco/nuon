package render

const (
	defaultPrefix string = "nuon"
)

func EnsurePrefix(data map[string]interface{}) map[string]interface{} {
	_, isPrefixed := data[defaultPrefix]
	if !isPrefixed {
		data = map[string]interface{}{
			defaultPrefix: data,
		}
	}

	return data
}
