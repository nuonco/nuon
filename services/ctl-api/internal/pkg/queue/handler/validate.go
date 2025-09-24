package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const ValidateUpdateName string = "validate"

const validateUpdateType = handlerTypeUpdate

type ValidateResponse struct{}

func (w *handler) validateHandler(ctx workflow.Context) (*ValidateResponse, error) {
	if w.sig == nil {
		return nil, errors.New("signal was empty can not proceed")
	}

	if err := w.sig.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "validate method failed")
	}

	return nil, nil
}
