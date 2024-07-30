package config

import jsoniter "github.com/json-iterator/go"

func ToJSON(obj interface{}) ([]byte, error) {
	json := jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		TagKey:                 "mapstructure",
	}.Froze()

	byts, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return byts, nil
}
