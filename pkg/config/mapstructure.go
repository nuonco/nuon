package config

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func newJSON() jsoniter.API {
	return jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		TagKey:                 "mapstructure",
	}.Froze()
}

func nestWithName(name string, obj map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		name: obj,
	}
}

// why: toMapStructure: this allows us to convert any type into a mapstructure, so we can easily work back and forth with
// what will ultimately become terraform json.
//
// we go from struct -> json -> mapstructure, because mitchellh/mapstructure doesn't have good support for going from
// struct to json, and we want to be able to use `omitempty`.
func toMapStructure[T any](input T) (map[string]interface{}, error) {
	json := newJSON()

	byts, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("unable to create json: %w", err)
	}

	var mapStruct map[string]interface{}
	if err := json.Unmarshal(byts, &mapStruct); err != nil {
		return nil, fmt.Errorf("unable to convert to mapstruct: %w", err)
	}

	return mapStruct, nil
}
