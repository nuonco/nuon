package createorg

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

func (w *wkflow) CreateOrg(ctx workflow.Context, req *jobsv1.CreateOrgRequest) (*jobsv1.CreateOrgResponse, error) {
	var resp TriggerJobResponse

	act := &activities{}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	fut := workflow.ExecuteActivity(ctx, act.TriggerJob, req.OrgId)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}

	return &jobsv1.CreateOrgResponse{}, nil
}
