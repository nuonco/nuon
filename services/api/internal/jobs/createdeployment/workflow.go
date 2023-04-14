package createdeployment

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
)

type CreateDeploymentRequest struct {
	DeploymentID string `validate:"required"`
}

type CreateDeploymentResponse struct{}

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateDeployment(ctx workflow.Context, req CreateDeploymentRequest) (CreateDeploymentResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create deployment")

	return CreateDeploymentResponse{}, nil
}
