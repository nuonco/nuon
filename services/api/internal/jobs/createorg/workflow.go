package createorg

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) CreateOrg(ctx workflow.Context, req *jobsv1.CreateOrgRequest) (*jobsv1.CreateOrgResponse, error) {
	act := &activities{}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var resp TriggerJobResponse
	fut := workflow.ExecuteActivity(ctx, act.TriggerOrgProvision, req.OrgId)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}

	shrdAct := &sharedactivities.Activities{}

	pollRequest := &activitiesv1.PollWorkflowRequest{
		Namespace:    "orgs",
		WorkflowName: "Signup",
		WorkflowId:   resp.WorkflowID,
	}

	// set poll timeout
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: sharedactivities.PollActivityTimeout * sharedactivities.MaxActivityRetries,
		StartToCloseTimeout:    sharedactivities.PollActivityTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: sharedactivities.MaxActivityRetries,
		},
	})

	var pollResp activitiesv1.PollWorkflowResponse
	fut = workflow.ExecuteActivity(ctx, shrdAct.PollWorkflow, pollRequest)
	err := fut.Get(ctx, &pollResp)
	if err != nil {
		return nil, fmt.Errorf("unable to poll for workflow response: %w", err)
	}

	return &jobsv1.CreateOrgResponse{}, nil
}
