package activities

import (
	"context"
)

type UpdateStatusRequest struct {
	InstallID         string `validate:"required"`
	Status            string `validate:"required"`
	StatusDescription string `validate:"required"`
}

// Deprecated: Status and StatusDescription fields have been removed from Install.
// This activity is retained for backward compatibility with in-flight workflows.
func (a *Activities) UpdateStatus(_ context.Context, _ UpdateStatusRequest) error {
	return nil
}
