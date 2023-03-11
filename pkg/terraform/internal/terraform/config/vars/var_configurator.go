package vars

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
)

type varConfigurator struct {
	M map[string]interface{} `validate:"required"`

	// internal state
	validator *validator.Validate
}

type varConfiguratorOption func(*varConfigurator) error

func New(v *validator.Validate, opts ...varConfiguratorOption) (*varConfigurator, error) {
	vc := &varConfigurator{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating s3 fetcher: validator is nil")
	}
	vc.validator = v

	for _, opt := range opts {
		if err := opt(vc); err != nil {
			return nil, err
		}
	}

	if err := vc.validator.Struct(vc); err != nil {
		return nil, err
	}

	return vc, nil
}

func WithVars(m map[string]interface{}) varConfiguratorOption {
	return func(vc *varConfigurator) error {
		vc.M = m
		return nil
	}
}

func (v *varConfigurator) JSON(w io.Writer) error {
	byts, err := json.Marshal(v.M)
	if err != nil {
		return err
	}

	_, err = w.Write(byts)
	return err
}
