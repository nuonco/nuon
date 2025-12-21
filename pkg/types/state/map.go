package state

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func AsMap(i any) (map[string]any, error) {
	byts, err := json.Marshal(i)
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to json")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(byts, &obj); err != nil {
		return nil, errors.Wrap(err, "unable to convert to map[string]interface{}")
	}

	return obj, nil
}
