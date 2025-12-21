package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const (
	CancelUpdateName string = "cancel"
	CancelUpdateType        = handlerTypeUpdate
)

var ErrAlreadyExecuted = errors.New("signal already succeeded")

type CancelRequest struct{}

type CancelResponse struct{}

func (h *handler) cancelHandler(ctx workflow.Context, req *CancelRequest) (*CancelResponse, error) {
	if h.finished {
		return nil, ErrAlreadyExecuted
	}

	h.canceled = true

	if h.executingCtx != nil {
		h.executingCancel()
	}

	return &CancelResponse{}, nil
}
