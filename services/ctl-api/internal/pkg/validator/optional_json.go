package validator

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

func optionalJSONValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if value == "" {
		return true
	}

	var js json.RawMessage
	return json.Unmarshal([]byte(value), &js) == nil
}
