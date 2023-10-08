package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationBuild            Operation = "build"
	OperationDelete           Operation = "delete"
	OperationPollDependencies Operation = "poll_dependencies"
)

type Signal struct {
	Operation Operation `validate:"required"`

	BuildID string `validate:"required_if=Operation build"`
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}
