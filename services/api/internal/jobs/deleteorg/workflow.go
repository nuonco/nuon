package deleteorg

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

func (w *wkflow) DeleteOrg(ctx workflow.Context, req *jobsv1.DeleteOrgRequest) (*jobsv1.DeleteOrgResponse, error) {
	var resp TriggerJobResponse

	act := &activities{}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	fut := workflow.ExecuteActivity(ctx, act.TriggerOrgJob, req.OrgId)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}

	return &jobsv1.DeleteOrgResponse{}, nil
}
