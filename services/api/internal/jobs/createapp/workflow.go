package createapp

import (
	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	"go.temporal.io/sdk/workflow"
)

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateApp(ctx workflow.Context, req jobsv1.CreateAppRequest) (jobsv1.CreateAppResponse, error) {
	l := workflow.GetLogger(ctx)
	l.Info("create app")
	return jobsv1.CreateAppResponse{}, nil
}
