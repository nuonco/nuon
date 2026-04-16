package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const (
	ReadyHandlerName string = "ready"
	ReadyHandlerType        = handlerTypeUpdate
)

type ReadyRequest struct{}

type ReadyResponse struct {
	RunID string
}

func (h *handler) readyHandler(ctx workflow.Context, req *ReadyRequest) (*ReadyResponse, error) {
	if err := workflow.Await(ctx, func() bool {
		return h.ready
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	return &ReadyResponse{}, nil
}
