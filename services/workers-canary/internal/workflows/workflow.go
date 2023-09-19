package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/go-playground/validator/v10"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
)

const (
	defaultStartActivityTimeout time.Duration = time.Second * 5
	defaultPollActivityTimeout  time.Duration = time.Minute * 30
	defaultMaxActivityRetries		  = 5
	defaultRegion				  = "us-west-2"
)

type wkflow struct {
	cfg  workers.Config
	acts *activities.Activities
	l    *zap.Logger
}

func New(v *validator.Validate, cfg workers.Config) (*wkflow, error) {
	return &wkflow{
		cfg: cfg,
		l:   zap.L(),
	}, nil
}

func (w *wkflow) startWorkflow(
	ctx workflow.Context,
	namespace, name string,
	msg protoreflect.ProtoMessage,
) (string, error) {
	l := workflow.GetLogger(ctx)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultStartActivityTimeout * defaultMaxActivityRetries,
		StartToCloseTimeout:	defaultStartActivityTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: defaultMaxActivityRetries,
		},
	})

	req, err := anypb.New(msg)
	if err != nil {
		return "", fmt.Errorf("unable to create any request: %w", err)
	}
	startReq := &activitiesv1.StartWorkflowRequest{
		Namespace:    namespace,
		WorkflowName: name,
		Request:      req,
	}

	var resp activitiesv1.StartWorkflowResponse
	fut := workflow.ExecuteActivity(ctx, "StartWorkflow", startReq)
	if err := fut.Get(ctx, &resp); err != nil {
		l.Info("unable to start workflow", zap.Error(err))
		return "", fmt.Errorf("unable to start workflow: %w", err)
	}
	l.Info("successfully started %s.%s workflow", namespace, name)

	return resp.WorkflowId, nil
}

func (w *wkflow) pollWorkflow(
	ctx workflow.Context,
	namespace, name, workflowID string,
) (*activitiesv1.PollWorkflowResponse, error) {
	l := workflow.GetLogger(ctx)

	pollReq := &activitiesv1.PollWorkflowRequest{
		Namespace:    namespace,
		WorkflowName: name,
		WorkflowId:   workflowID,
	}

	shrdAct := &sharedactivities.Activities{}
	// set poll timeout
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: sharedactivities.PollActivityTimeout * sharedactivities.MaxActivityRetries,
		StartToCloseTimeout:	sharedactivities.PollActivityTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: sharedactivities.MaxActivityRetries,
		},
	})

	var pollResp activitiesv1.PollWorkflowResponse
	fut := workflow.ExecuteActivity(ctx, shrdAct.PollWorkflow, pollReq)
	err := fut.Get(ctx, &pollResp)
	if err != nil {
		return nil, fmt.Errorf("unable to poll for workflow response: %w", err)
	}

	resp, wkflowErr := anypb.New(&pollResp)
	if wkflowErr != nil {
		return nil, fmt.Errorf("unable to get response: %w", wkflowErr)
	}

	l.Info("successfully got %s.%s workflow response", namespace, name)
	return &activitiesv1.PollWorkflowResponse{
		Response: resp,
	}, nil
}
