package createdeployment

import (
	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	"go.temporal.io/sdk/workflow"
)

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateDeployment(ctx workflow.Context, req jobsv1.CreateDeploymentRequest) (jobsv1.CreateDeploymentResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create deployment")

	return jobsv1.CreateDeploymentResponse{}, nil
}
