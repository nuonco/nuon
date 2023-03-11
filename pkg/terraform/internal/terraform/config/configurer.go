package config

import (
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
)

type writeFactory interface {
	GetWriter(string) (io.WriteCloser, error)
}

type configurator interface {
	JSON(io.Writer) error
}

type configurer struct {
	WriteFactory  writeFactory            `validate:"required"`
	Configurators map[string]configurator `validate:"required,gt=0,dive,required"`

	// internal state
	validator *validator.Validate
}

type configurerOptions func(*configurer) error

func New(v *validator.Validate, opts ...configurerOptions) (*configurer, error) {
	c := &configurer{Configurators: make(map[string]configurator)}

	if v == nil {
		return nil, fmt.Errorf("error instantiating configurator: validator is nil")
	}
	c.validator = v

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if err := c.validator.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}

func WithConfigurator(name string, ctr configurator) configurerOptions {
	return func(c *configurer) error {
		c.Configurators[name] = ctr
		return nil
	}
}

func WithWriteFactory(w writeFactory) configurerOptions {
	return func(c *configurer) error {
		c.WriteFactory = w
		return nil
	}
}

func (c *configurer) Configure() error {
	for name, ctr := range c.Configurators {
		if err := c.configure(name, ctr); err != nil {
			return err
		}
	}
	return nil
}

func (c *configurer) configure(name string, ctr configurator) error {
	w, err := c.WriteFactory.GetWriter(name)
	if err != nil {
		return err
	}
	if w == nil {
		return fmt.Errorf("somehow ended up with nil writer. this shouldn't happen")
	}
	defer w.Close()
	return ctr.JSON(w)
}
