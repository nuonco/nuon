package provision

import (
	"fmt"
	"time"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	kmsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/kms/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/runner/v1"
	serverv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/server/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/iam"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/kms"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/runner"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/server"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) Provision(ctx workflow.Context, req *orgsv1.ProvisionRequest) (*orgsv1.ProvisionResponse, error) {
	resp := &orgsv1.ProvisionResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return resp, err
	}

	l.Debug("provisioning iam for org")
	iamResp, err := execProvisionIAMWorkflow(ctx, w.cfg, &iamv1.ProvisionIAMRequest{
		OrgId:       req.OrgId,
		Reprovision: req.Reprovision,
	})
	if err != nil {
		err = fmt.Errorf("failed to provision iam: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	resp.IamRoles = iamResp

	l.Debug("provisioning kms for org")
	kmsResp, err := execProvisionKMSWorkflow(ctx, w.cfg, &kmsv1.ProvisionKMSRequest{
		OrgId:             req.OrgId,
		SecretsIamRoleArn: iamResp.SecretsRoleArn,
		Reprovision:       req.Reprovision,
	})
	if err != nil {
		err = fmt.Errorf("failed to provision kms: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	resp.Kms = kmsResp

	l.Debug("provisioning waypoint org server")
	serverResp, err := execProvisionWaypointServerWorkflow(ctx, w.cfg, &serverv1.ProvisionServerRequest{
		OrgId:       req.OrgId,
		Region:      req.Region,
		Reprovision: req.Reprovision,
		CustomCert:  req.CustomCert,
	})
	if err != nil {
		err = fmt.Errorf("failed to install server: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	resp.Server = serverResp

	l.Debug("installing waypoint org runner")
	_, err = execInstallWaypointRunnerWorkflow(ctx, w.cfg, &runnerv1.ProvisionRunnerRequest{
		OrgId:         req.OrgId,
		OdrIamRoleArn: iamResp.OdrRoleArn,
		Region:        req.Region,
		Reprovision:   req.Reprovision,
	})
	if err != nil {
		err = fmt.Errorf("failed to install runner: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	w.finishWorkflow(ctx, req, resp, err)
	l.Debug("finished signup", "response", resp)
	return resp, nil
}

func execProvisionWaypointServerWorkflow(
	ctx workflow.Context,
	cfg workers.Config,
	req *serverv1.ProvisionServerRequest,
) (*serverv1.ProvisionServerResponse, error) {
	var resp *serverv1.ProvisionServerResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing install waypoint workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-server", req.OrgId),
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := server.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionServer, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execInstallWaypointRunnerWorkflow(
	ctx workflow.Context,
	cfg workers.Config,
	iwrr *runnerv1.ProvisionRunnerRequest,
) (*runnerv1.ProvisionRunnerResponse, error) {
	var resp runnerv1.ProvisionRunnerResponse

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-runner", iwrr.OrgId),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := runner.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionRunner, iwrr)

	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, err
	}

	return &resp, nil
}

func execProvisionIAMWorkflow(
	ctx workflow.Context,
	cfg workers.Config,
	req *iamv1.ProvisionIAMRequest,
) (*iamv1.ProvisionIAMResponse, error) {
	var resp iamv1.ProvisionIAMResponse

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-iam", req.OrgId),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := iam.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionIAM, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, err
	}

	return &resp, nil
}

func execProvisionKMSWorkflow(
	ctx workflow.Context,
	cfg workers.Config,
	req *kmsv1.ProvisionKMSRequest,
) (*kmsv1.ProvisionKMSResponse, error) {
	var resp kmsv1.ProvisionKMSResponse

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-kms", req.OrgId),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := kms.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionKMS, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, err
	}

	return &resp, nil
}
