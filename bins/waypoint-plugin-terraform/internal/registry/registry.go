package registry

import "github.com/go-playground/validator/v10"

func New() *Registry {
	return &Registry{
		v: validator.New(),
	}
}

type Registry struct {
	config Config
	v      *validator.Validate
}
