package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/mitchellh/mapstructure"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/sandbox"
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
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.SandboxInput{
				Type:      planv1.SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION,
				OrgId:     req.OrgId,
				AppId:     req.AppId,
				InstallId: req.InstallId,
				SandboxSettings: &planv1.SandboxSettings{
					Name:    req.SandboxSettings.Name,
					Version: req.SandboxSettings.Version},
				AccountSettings: &planv1.SandboxInput_Aws{
					Aws: &planv1.AWSSettings{
						Region:    req.AccountSettings.Region,
						AccountId: req.AccountSettings.AwsAccountId,
						RoleArn:   req.AccountSettings.AwsRoleArn,
					},
				},
			},
		},
	}

	spr, err := sandbox.Plan(ctx, &cpReq)
	if err != nil {
		err = fmt.Errorf("unable to plan sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	if req.PlanOnly {
		l.Info("skipping the rest of the workflow - plan only")
		w.finishWorkflow(ctx, req, resp, nil)
		return resp, nil
	}

	seReq := executev1.ExecutePlanRequest{Plan: spr.Plan}
	ser, err := sandbox.Execute(ctx, &seReq)
	if err != nil {
		err = fmt.Errorf("unable to execute sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	tfOutputs, err := sandbox.ParseTerraformOutputs(ser.GetTerraformOutputs())
	if err != nil {
		err = fmt.Errorf("invalid sandbox outputs: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	// convert terraform outputs to map and add to response
	tfOutputsMap, err := toMap(tfOutputs)
	if err != nil {
		err = fmt.Errorf("unable to decode to stringmap: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	resp.TerraformOutputs = tfOutputsMap

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

func toMap(tfOutputs sandbox.TerraformOutputs) (map[string]string, error) {
	var output map[string]string
	if err := mapstructure.Decode(tfOutputs, &output); err != nil {
		return nil, err
	}

	return output, nil
}
