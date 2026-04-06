package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const (
	WakeUpdateName string = "wake"
	WakeUpdateType        = handlerTypeUpdate
)

type WakeResponse struct{}

func (h *handler) wakeHandler(ctx workflow.Context) (*WakeResponse, error) {
	if !h.sleeping {
		return nil, errors.New("handler is not sleeping")
	}
	h.woken = true
	h.sleeping = false
	return &WakeResponse{}, nil
}
