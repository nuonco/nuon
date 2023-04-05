package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/go-playground/validator/v10"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1/activities/v1"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
	defaultRegion          = "us-west-2"
)

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type wkflow struct {
	cfg workers.Config
}

func New(v *validator.Validate, cfg workers.Config) (*wkflow, error) {
	return &wkflow{
		cfg: cfg,
	}, nil
}

func (w *wkflow) startWorkflow(ctx workflow.Context, namespace, name string, msg protoreflect.ProtoMessage) (string, error) {
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
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
	fut := workflow.ExecuteActivity(ctx, activitiesv1.Activity_ACTIVITY_START_WORKFLOW.String(), startReq)
	if err := fut.Get(ctx, &resp); err != nil {
		return "", fmt.Errorf("unable to get start workflow response: %w", err)
	}
	l.Info("successfully started %s.%s workflow", namespace, name)

	return resp.WorkflowId, nil
}

func (w *wkflow) pollWorkflow(ctx workflow.Context, namespace, name, workflowID string) (*activitiesv1.PollWorkflowResponse, error) {
	l := workflow.GetLogger(ctx)
	pollReq := &activitiesv1.PollWorkflowRequest{
		Namespace:    namespace,
		WorkflowName: name,
		WorkflowId:   workflowID,
	}

	var resp activitiesv1.PollWorkflowResponse
	ctx = configureActivityOptions(ctx)
	fut := workflow.ExecuteActivity(ctx, activitiesv1.Activity_ACTIVITY_POLL_WORKFLOW.String(), pollReq)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get poll workflow response: %w", err)
	}
	l.Info("successfully got %s.%s workflow response", namespace, name)

	return &resp, nil
}

type step struct {
	name string
	fn   func(workflow.Context, string, *canaryv1.ProvisionRequest) (*canaryv1.Step, error)
}
