package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	FinishedHandlerName string = "finished"
	FinishedHandlerType        = handlerTypeUpdate
)

type FinishedRequest struct{}

type FinishedResponse struct {
	Status            app.Status `json:"status"`
	StatusDescription string     `json:"status_description,omitempty"`
}

func (h *handler) finishedHandler(ctx workflow.Context, req *FinishedRequest) (*FinishedResponse, error) {
	if err := workflow.Await(ctx, func() bool {
		return h.finished
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

	return &FinishedResponse{
		Status:            h.finishedStatus,
		StatusDescription: h.finishedErr,
	}, nil
}
