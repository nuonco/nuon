package createinstall

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

func (w *wkflow) CreateInstall(ctx workflow.Context, req *jobsv1.CreateInstallRequest) (*jobsv1.CreateInstallResponse, error) {
	var resp TriggerJobResponse

	act := &activities{}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	fut := workflow.ExecuteActivity(ctx, act.TriggerInstallJob, req.InstallId)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}

	shrdAct := &sharedactivities.Activities{}

	pollRequest := &activitiesv1.PollWorkflowRequest{
		Namespace:    "installs",
		WorkflowName: "Provision",
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

	return &jobsv1.CreateInstallResponse{}, nil
}
