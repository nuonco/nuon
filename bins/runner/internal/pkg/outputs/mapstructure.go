package outputs

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func ToMapstructure(obj interface{}) (map[string]interface{}, error) {
	byts, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create json")
	}

	var out map[string]interface{}
	if err := json.Unmarshal(byts, &out); err != nil {
		return nil, errors.Wrap(err, "unable to convert to map[string]interface{}")
	}

	return out, nil
}
