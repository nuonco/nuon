package deprovision

import (
	"fmt"
	"time"

	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/activities"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/runner"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/structpb"
)

func (w *wkflow) execFetchSandboxOutputs(
	ctx workflow.Context,
	req activities.FetchSandboxOutputsRequest,
) (*structpb.Struct, error) {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var resp *structpb.Struct
	fut := workflow.ExecuteActivity(ctx, w.sharedActs.FetchSandboxOutputs, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func execProvisionRunner(
	ctx workflow.Context,
	cfg workers.Config,
	iwrr *runnerv1.ProvisionRunnerRequest,
) (*runnerv1.ProvisionRunnerResponse, error) {
	resp := &runnerv1.ProvisionRunnerResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-runner", iwrr.InstallId),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := runner.NewWorkflow(nil, cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionRunner, iwrr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execDeprovisionRunner(
	ctx workflow.Context,
	cfg workers.Config,
	iwrr *runnerv1.DeprovisionRunnerRequest,
) (*runnerv1.DeprovisionRunnerResponse, error) {
	resp := &runnerv1.DeprovisionRunnerResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-deprovision-runner", iwrr.InstallId),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := runner.NewWorkflow(nil, cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.DeprovisionRunner, iwrr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
