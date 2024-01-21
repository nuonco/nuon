package provision

import (
	"fmt"
	"time"

	dnsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/dns/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/activities"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/dns"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/runner"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/structpb"
)

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

func execProvisionDNS(
	ctx workflow.Context,
	cfg workers.Config,
	req *dnsv1.ProvisionDNSRequest,
	installID string,
) (*dnsv1.ProvisionDNSResponse, error) {
	resp := &dnsv1.ProvisionDNSResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-dns", installID),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := dns.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionDNS, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execCheckIAMRole(
	ctx workflow.Context,
	act *Activities,
	req CheckIAMRoleRequest,
) error {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	var resp CheckIAMRoleResponse
	fut := workflow.ExecuteActivity(ctx, act.CheckIAMRole, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}

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
