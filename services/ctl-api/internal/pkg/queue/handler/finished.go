package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const (
	FinishedHandlerName string = "finished"
	FinishedHandlerType        = handlerTypeUpdate
)

type FinishedRequest struct{}

type FinishedResponse struct{}

func (w *handler) finishedHandler(ctx workflow.Context, req *FinishedRequest) (*FinishedResponse, error) {
	if err := workflow.Await(ctx, func() bool {
		return w.finished
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	return &FinishedResponse{}, nil
}
