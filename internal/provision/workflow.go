package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	runnerv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/runner/v1"
	sandboxv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/sandbox/v1"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/workers-installs/internal/provision/sandbox"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// Provision is a workflow that creates an app install sandbox using terraform
func (w wkflow) Provision(ctx workflow.Context, req *installsv1.ProvisionRequest) (*installsv1.ProvisionResponse, error) {
	resp := &installsv1.ProvisionResponse{}
	l := workflow.GetLogger(ctx)

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return resp, err
	}

	psReq := &sandboxv1.ProvisionSandboxRequest{
		AppId:           req.AppId,
		OrgId:           req.OrgId,
		InstallId:       req.InstallId,
		AccountSettings: req.AccountSettings,
		SandboxSettings: req.SandboxSettings,
	}
	psr, err := execProvisionSandbox(ctx, w.cfg, psReq)
	if err != nil {
		err = fmt.Errorf("unable to provision sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	resp.TerraformOutputs = psr.TerraformOutputs

	tfOutputs, err := sandbox.ParseTerraformOutputs(psr.TerraformOutputs)
	if err != nil {
		err = fmt.Errorf("invalid sandbox outputs: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	prReq := &runnerv1.ProvisionRunnerRequest{
		OrgId:         req.OrgId,
		AppId:         req.AppId,
		InstallId:     req.InstallId,
		OdrIamRoleArn: tfOutputs.OdrIAMRoleArn,
		Region:        req.AccountSettings.Region,
		ClusterInfo: &runnerv1.KubeClusterInfo{
			Id:             tfOutputs.ClusterID,
			Endpoint:       tfOutputs.ClusterEndpoint,
			CaData:         tfOutputs.ClusterCA,
			TrustedRoleArn: w.cfg.NuonAccessRoleArn,
		},
	}
	if _, err = execProvisionRunner(ctx, w.cfg, prReq); err != nil {
		err = fmt.Errorf("unable to provision install runner: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	l.Debug("finished provisioning", "response", resp)
	w.finishWorkflow(ctx, req, resp, nil)
	return resp, nil
}

func execProvisionSandbox(
	ctx workflow.Context,
	cfg workers.Config,
	iwrr *sandboxv1.ProvisionSandboxRequest,
) (*sandboxv1.ProvisionSandboxResponse, error) {
	resp := &sandboxv1.ProvisionSandboxResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 30,
		WorkflowTaskTimeout:      time.Minute * 15,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := sandbox.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionSandbox, iwrr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
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
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := runner.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionRunner, iwrr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
