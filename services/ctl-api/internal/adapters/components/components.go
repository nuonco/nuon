package components

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Adapter struct {
	cfg *internal.Config
}

func New(v *validator.Validate, cfg *internal.Config) (*Adapter, error) {
	return &Adapter{
		cfg: cfg,
	}, nil
}
