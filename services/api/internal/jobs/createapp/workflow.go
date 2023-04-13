package createapp

import "context"

type CreateAppRequest struct {
	AppID string `validate:"required"`
}

type CreateAppResponse struct{}

type workflow struct{}

func (w *workflow) CreateApp(ctx context.Context, req CreateAppRequest) (CreateAppResponse, error) {
	return CreateAppResponse{}, nil
}
