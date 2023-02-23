package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1"
	runnerv1 "github.com/powertoolsdev/protos/workflows/generated/types/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/outputs"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
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

	cpReq := planv1.CreatePlanRequest{
		Type: planv1.PlanType_PLAN_TYPE_TERRAFORM_SANDBOX,
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.Sandbox{
				OrgId:           req.OrgId,
				AppId:           req.AppId,
				InstallId:       req.InstallId,
				SandboxSettings: &planv1.SandboxSettings{Name: req.SandboxSettings.Name, Version: req.SandboxSettings.Version},
				// TODO(jdt): accept this from the API and set it here?
				// it's defaulted in workers-executors so no need to double hard-code
				// TerraformVersion: new(string),
				RunType: planv1.TerraformRunType_TERRAFORM_RUN_TYPE_APPLY,
				AccountSettings: &planv1.Sandbox_Aws{
					Aws: &planv1.AWSSettings{
						Region:    req.AccountSettings.Region,
						AccountId: req.AccountSettings.AwsAccountId,
						RoleArn:   req.AccountSettings.AwsRoleArn,
					},
				},
			},
		},
	}

	if req.PlanOnly {
		l.Info("skipping the rest of the workflow - plan only")
		w.finishWorkflow(ctx, req, resp, nil)
		return resp, nil
	}

	spr, err := execSandboxPlan(ctx, &cpReq)
	if err != nil {
		err = fmt.Errorf("unable to plan sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	seReq := executev1.ExecutePlanRequest{Plan: spr.Plan}
	ser, err := execSandboxExecute(ctx, &seReq)
	if err != nil {
		err = fmt.Errorf("unable to execute sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	tfOutputs, err := outputs.ParseTerraformOutputs(ser.GetTerraformOutputs())
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

func execSandboxPlan(
	ctx workflow.Context,
	cpr *planv1.CreatePlanRequest,
) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 30,
		WorkflowTaskTimeout:      time.Minute * 15,
		TaskQueue:                "executors",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", cpr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execSandboxExecute(
	ctx workflow.Context,
	cpr *executev1.ExecutePlanRequest,
) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 30,
		WorkflowTaskTimeout:      time.Minute * 15,
		TaskQueue:                "executors",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", cpr)

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
