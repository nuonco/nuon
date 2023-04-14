package createinstall

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"
)

type CreateInstallRequest struct {
	InstallID string `validate:"required"`
}

type CreateInstallResponse struct{}

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateInstall(ctx workflow.Context, req CreateInstallRequest) (CreateInstallResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create install")
	return CreateInstallResponse{}, nil
}
