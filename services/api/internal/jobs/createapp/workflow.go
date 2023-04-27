package createapp

import (
	"time"

	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	"go.temporal.io/sdk/workflow"
)

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateApp(ctx workflow.Context, req *jobsv1.CreateAppRequest) (*jobsv1.CreateAppResponse, error) {
	var resp jobsv1.CreateAppResponse

	act := &activities{}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	fut := workflow.ExecuteActivity(ctx, act.TriggerAppJob, req.AppId)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
