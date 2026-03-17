package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (h *handler) validateHandler(ctx workflow.Context) (*ValidateResponse, error) {
	if h.validated {
		// Already validated; this is a no-op (idempotent on queue restart).
		return &ValidateResponse{}, nil
	}

	if h.sig == nil {
		return nil, errors.New("signal was empty can not proceed")
	}

	if err := h.sig.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "validate method failed")
	}

	h.validated = true
	return &ValidateResponse{}, nil
}
