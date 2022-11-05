package workers

import "github.com/go-playground/validator/v10"

type Config struct {
	Value string `config:"value" validate:"required"`
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}
