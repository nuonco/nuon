package createapp

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
)

type CreateAppRequest struct {
	AppID string `validate:"required"`
}

type CreateAppResponse struct{}

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateApp(ctx workflow.Context, req CreateAppRequest) (CreateAppResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create app")
	return CreateAppResponse{}, nil
}
