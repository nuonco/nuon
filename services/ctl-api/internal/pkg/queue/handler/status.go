package handler

import (
	"go.temporal.io/sdk/workflow"
)

const (
	StatusQueryName string = "status"
	StatusQueryType        = handlerTypeQuery
)

type StatusRequest struct{}

type StatusResponse struct {
	Finished bool
	Canceled bool
}

func (h *handler) statusHandler(ctx workflow.Context, req *StatusRequest) (*StatusResponse, error) {
	return &StatusResponse{
		Finished: h.finished,
		Canceled: h.canceled,
	}, nil
}
