package createinstall

import (
	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	"go.temporal.io/sdk/workflow"
)

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateInstall(ctx workflow.Context, req jobsv1.CreateInstallRequest) (jobsv1.CreateInstallResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create install")
	return jobsv1.CreateInstallResponse{}, nil
}
