package createdeployment

import "context"

type CreateDeploymentRequest struct {
	DeploymentID string `validate:"required"`
}

type CreateDeploymentResponse struct{}

type workflow struct{}

func (w *workflow) CreateDeployment(ctx context.Context, req CreateDeploymentRequest) (CreateDeploymentResponse, error) {
	return CreateDeploymentResponse{}, nil
}
