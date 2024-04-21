package worker

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Operation string

const (
	OperationProvision     Operation = "provision"
	OperationDelete        Operation = "delete"
	OperationForceDelete   Operation = "force_delete"
	OperationDeprovision   Operation = "deprovision"
	OperationReprovision   Operation = "reprovision"
	OperationRestart       Operation = "restart"
	OperationInviteCreated Operation = "invite_created"
)

type Signal struct {
	Operation Operation `validate:"required"`

	// for InviteCreated event
	Email string
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}
