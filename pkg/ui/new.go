package ui

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type logger struct {
	v *validator.Validate

	Zap    *zap.Logger
	Silent bool
	JSON   bool
}

type loggerOption func(*logger) error

func New(v *validator.Validate, opts ...loggerOption) (*logger, error) {
	log := &logger{
		v:   v,
		Zap: zap.L(),
	}
	for idx, opt := range opts {
		if err := opt(log); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := log.v.Struct(log); err != nil {
		return nil, fmt.Errorf("unable to validate command: %w", err)
	}

	return log, nil
}

// WithSilent prevents all output
func WithSilent(silent bool) loggerOption {
	return func(l *logger) error {
		l.Silent = silent
		return nil
	}
}

// WithJSON outputs everything as json
func WithJSON(enableJSON bool) loggerOption {
	return func(l *logger) error {
		l.JSON = enableJSON
		return nil
	}
}
