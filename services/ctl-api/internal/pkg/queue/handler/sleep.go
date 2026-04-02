package handler

import (
	"go.temporal.io/sdk/workflow"
)

const (
	SleepUpdateName string = "sleep"
	SleepUpdateType        = handlerTypeUpdate
)

type SleepResponse struct{}

func (h *handler) sleepHandler(ctx workflow.Context) (*SleepResponse, error) {
	h.sleeping = true
	h.stopped = true
	return &SleepResponse{}, nil
}
