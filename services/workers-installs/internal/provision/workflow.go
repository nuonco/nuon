package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	awseks "github.com/powertoolsdev/mono/pkg/sandboxes/aws-eks"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	dnsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/dns/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/dns"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/runner"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/sandbox"
	enumspb "go.temporal.io/api/enums/v1"
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

	runTyp := planv1.SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION
	if req.PlanOnly {
		runTyp = planv1.SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION_PLAN
	}

	cpReq := planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.SandboxInput{
				Type:             runTyp,
				OrgId:            req.OrgId,
				AppId:            req.AppId,
				InstallId:        req.InstallId,
				TerraformVersion: req.SandboxSettings.TerraformVersion,
				SandboxSettings: &planv1.SandboxSettings{
					Name:    req.SandboxSettings.Name,
					Version: req.SandboxSettings.Version,
				},
				AccountSettings: &planv1.SandboxInput_Aws{
					Aws: &planv1.AWSSettings{
						Region:    req.AccountSettings.Region,
						AccountId: req.AccountSettings.AwsAccountId,
						RoleArn:   req.AccountSettings.AwsRoleArn,
					},
				},
				RootDomain: fmt.Sprintf("%s.%s", req.InstallId, w.cfg.PublicDomain),
			},
		},
	}

	spr, err := sandbox.Plan(ctx, &cpReq)
	if err != nil {
		err = fmt.Errorf("unable to plan sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	seReq := executev1.ExecutePlanRequest{Plan: spr.Plan}
	ser, err := sandbox.Execute(ctx, &seReq)
	if err != nil {
		err = fmt.Errorf("unable to execute sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}
	if req.PlanOnly {
		l.Info("skipping the rest of the workflow - plan only")
		w.finishWorkflow(ctx, req, resp, nil)
		return resp, nil
	}

	resp.TerraformOutputs = ser.GetTerraformOutputs()
	tfOutputs, err := awseks.ParseTerraformOutputs(ser.GetTerraformOutputs())
	if err != nil {
		err = fmt.Errorf("invalid sandbox outputs: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	dnsReq := &dnsv1.ProvisionDNSRequest{
		Domain:      tfOutputs.PublicDomain.Name,
		Nameservers: awseks.ToStringSlice(tfOutputs.PublicDomain.Nameservers),
	}
	_, err = execProvisionDNS(ctx, w.cfg, dnsReq, req.InstallId)
	if err != nil {
		err = fmt.Errorf("unable to provision dns: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	prReq := &runnerv1.ProvisionRunnerRequest{
		OrgId:         req.OrgId,
		AppId:         req.AppId,
		InstallId:     req.InstallId,
		OdrIamRoleArn: tfOutputs.Runner.DefaultIAMRoleARN,
		Region:        req.AccountSettings.Region,
		ClusterInfo: &runnerv1.KubeClusterInfo{
			Id:             tfOutputs.Cluster.Name,
			Endpoint:       tfOutputs.Cluster.Endpoint,
			CaData:         tfOutputs.Cluster.CertificateAuthorityData,
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
		WorkflowID:               fmt.Sprintf("%s-provision-runner", iwrr.InstallId),
		WorkflowExecutionTimeout: time.Minute * 10,
		WorkflowTaskTimeout:      time.Minute * 5,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := runner.NewWorkflow(cfg)
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
