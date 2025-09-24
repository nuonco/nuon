package queue

import (
	"go.temporal.io/sdk/workflow"
)

const (
	StatusHandlerName string = "status"
	StatusHandlerType        = handlerTypeUpdate
)

type StatusRequest struct{}

type StatusResponse struct {
	Ready bool

	QueueDepthCount int
	InFlightCount   int
	InFlight        []string
}

func (w *queue) statusHandler(ctx workflow.Context, req *StatusRequest) (*StatusResponse, error) {
	resp := &StatusResponse{
		Ready: w.ready,
	}
	if !w.ready {
		return resp, nil
	}

	// fetch the entire status

	return nil, nil
}
