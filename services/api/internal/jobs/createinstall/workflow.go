package createinstall

import "context"

type CreateInstallRequest struct {
	InstallID string `validate:"required"`
}

type CreateInstallResponse struct{}

type workflow struct{}

func (w *workflow) CreateInstall(ctx context.Context, req CreateInstallRequest) (CreateInstallResponse, error) {
	return CreateInstallResponse{}, nil
}
