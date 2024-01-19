package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	awseks "github.com/powertoolsdev/mono/pkg/sandboxes/aws-eks"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	dnsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/dns/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
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

func (w wkflow) createPlanRequest(runTyp planv1.SandboxInputType, req *installsv1.ProvisionRequest) *planv1.CreatePlanRequest {
	return &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Sandbox{
			Sandbox: &planv1.SandboxInput{
				OrgId:           req.OrgId,
				AppId:           req.AppId,
				InstallId:       req.InstallId,
				RunId:           req.RunId,
				Type:            runTyp,
				SandboxSettings: req.SandboxSettings,
				AccountSettings: req.AccountSettings,
			},
		},
	}

}

func (w wkflow) executorsWorkflowID(req *installsv1.ProvisionRequest, jobName string) string {
	return fmt.Sprintf("%s-%s", req.RunId, jobName)
}

func (w wkflow) provisionNoopBuild(ctx workflow.Context, req *installsv1.ProvisionRequest) error {
	planReq := w.createPlanRequest(planv1.SandboxInputType_SANDBOX_INPUT_TYPE_NOOP_BUILD, req)
	planWorkflowID := w.executorsWorkflowID(req, "noop-build-plan")
	planResp, err := sandbox.Plan(ctx, planWorkflowID, planReq)
	if err != nil {
		return fmt.Errorf("unable to create noop-build plan: %w", err)
	}

	executeWorkflowID := w.executorsWorkflowID(req, "noop-build-execute")
	_, err = sandbox.Execute(ctx, executeWorkflowID,
		&executev1.ExecutePlanRequest{
			Plan: planResp.Plan,
		})
	if err != nil {
		return fmt.Errorf("unable to execute noop-build plan: %w", err)
	}

	return nil
}

func (w wkflow) provisionSandbox(ctx workflow.Context, req *installsv1.ProvisionRequest) (*executev1.ExecutePlanResponse, error) {
	runTyp := planv1.SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION
	if req.PlanOnly {
		runTyp = planv1.SandboxInputType_SANDBOX_INPUT_TYPE_PROVISION_PLAN
	}

	planReq := w.createPlanRequest(runTyp, req)
	planWorkflowID := w.executorsWorkflowID(req, "provision-plan")
	planResp, err := sandbox.Plan(ctx, planWorkflowID, planReq)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	executeWorkflowID := w.executorsWorkflowID(req, "provision-execute")
	execResp, err := sandbox.Execute(ctx, executeWorkflowID, &executev1.ExecutePlanRequest{
		Plan: planResp.Plan,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to execute plan: %w", err)
	}

	return execResp, nil
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
	act := NewActivities(nil, workers.Config{}, nil)

	if err := w.startWorkflow(ctx, req); err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return resp, err
	}

	if err := execCheckIAMRole(ctx, act, CheckIAMRoleRequest{
		RoleARN: req.AccountSettings.AwsRoleArn,
	}); err != nil {
		err = fmt.Errorf("unable to validate IAM role: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	if req.PlanOnly {
		l.Info("skipping the rest of the workflow - plan only")
		w.finishWorkflow(ctx, req, resp, nil)
		return resp, nil
	}

	if err := w.provisionNoopBuild(ctx, req); err != nil {
		err = fmt.Errorf("unable to create noop build: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	_, err := w.provisionSandbox(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to provision sandbox: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, err
	}

	outputs, err := execFetchSandboxOutputs(ctx, act, FetchSandboxOutputsRequest{
		OrgID:     req.OrgId,
		AppID:     req.AppId,
		InstallID: req.InstallId,
	})
	if err != nil {
		err = fmt.Errorf("unable to fetch sandbox outputs: %w", err)
		w.finishWorkflow(ctx, req, resp, err)
		return resp, nil
	}

	tfOutputs, err := awseks.ParseTerraformOutputs(outputs)
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
		RunnerType:    req.RunnerType,
	}

	if req.RunnerType == installsv1.RunnerType_RUNNER_TYPE_AWS_ECS {
		prReq.EcsClusterInfo = &runnerv1.ECSClusterInfo{
			ClusterArn:        tfOutputs.ECSCluster.ARN,
			InstallIamRoleArn: tfOutputs.Runner.InstallIAMRoleARN,
			RunnerIamRoleArn:  tfOutputs.Runner.RunnerIAMRoleARN,
			OdrIamRoleArn:     tfOutputs.Runner.ODRIAMRoleARN,
			VpcId:             tfOutputs.VPC.ID,
			SubnetIds:         generics.ToStringSlice(tfOutputs.VPC.PublicSubnetIDs),
			SecurityGroupId:   tfOutputs.VPC.DefaultSecurityGroupID,
		}
		prReq.OdrIamRoleArn = tfOutputs.Runner.ODRIAMRoleARN
	} else {
		prReq.EksClusterInfo = &runnerv1.KubeClusterInfo{
			Id:             tfOutputs.Cluster.Name,
			Endpoint:       tfOutputs.Cluster.Endpoint,
			CaData:         tfOutputs.Cluster.CertificateAuthorityData,
			TrustedRoleArn: w.cfg.NuonAccessRoleArn,
		}
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
