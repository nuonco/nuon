package build

import (
	"github.com/go-playground/validator/v10"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"go.temporal.io/sdk/workflow"
)

func New(v *validator.Validate) *wkflow {
	return &wkflow{}
}

type wkflow struct{}

func (w *wkflow) Build(ctx workflow.Context, req *buildv1.StartBuildRequest) (*buildv1.StartBuildResponse, error) {
	resp := buildv1.StartBuildResponse{}
	/*
		act := &activities{}

		activityOpts := workflow.ActivityOptions{
			ScheduleToCloseTimeout: time.Second * 5,
		}

		ctx = workflow.WithActivityOptions(ctx, activityOpts)

		var triggerResp TriggerJobResponse
		fut := workflow.ExecuteActivity(ctx, act.TriggerAppJob, req.AppId)
		if err := fut.Get(ctx, &triggerResp); err != nil {
			return nil, fmt.Errorf("unable to trigger workflow response: %w", err)
		}

		shrdAct := &sharedactivities.Activities{}

		pollRequest := &activitiesv1.PollWorkflowRequest{
			Namespace:    "apps",
			WorkflowName: "Provision",
			WorkflowId:   triggerResp.WorkflowID,
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
	*/
	return &resp, nil
}
