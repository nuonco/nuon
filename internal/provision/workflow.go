package provision

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/shortid"
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

	// parse IDs into short IDs, and use them for all subsequent requests
	shortIDs, err := shortid.ParseStrings(req.OrgId, req.AppId, req.InstallId)
	if err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}
	orgID, appID, installID := shortIDs[0], shortIDs[1], shortIDs[2]
	req.OrgId = orgID
	req.AppId = appID
	req.InstallId = installID

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// NOTE(jdt): this is just so that we can use the method names
	// the actual struct isn't used by temporal during dispatch at all
	act := NewProvisionActivities(workers.Config{}, nil)

	sReq := StartWorkflowRequest{
		AppID:                         appID,
		OrgID:                         orgID,
		InstallID:                     installID,
		InstallationsBucket:           w.cfg.InstallationsBucket,
		InstallationsAccessIAMRoleARN: fmt.Sprintf(w.cfg.OrgInstallationsRoleTemplate, orgID),
		ProvisionRequest:              req,
	}
	_, err = execStartWorkflow(ctx, act, sReq)
	if err != nil {
		err = fmt.Errorf("failed to execute start workflow activity: %w", err)
		w.finishWithErr(ctx, req, act, "start_workflow", err)
		return resp, err
	}

	psReq := &sandboxv1.ProvisionSandboxRequest{
		AppId:           appID,
		OrgId:           orgID,
		InstallId:       installID,
		AccountSettings: req.AccountSettings,
		SandboxSettings: req.SandboxSettings,
	}
	psr, err := execProvisionSandbox(ctx, w.cfg, psReq)
	if err != nil {
		err = fmt.Errorf("unable to provision sandbox: %w", err)
		w.finishWithErr(ctx, req, act, "provision_sandbox", err)
		return resp, err
	}
	resp.TerraformOutputs = psr.TerraformOutputs

	tfOutputs, err := sandbox.ParseTerraformOutputs(psr.TerraformOutputs)
	if err != nil {
		err = fmt.Errorf("invalid sandbox outputs: %w", err)
		w.finishWithErr(ctx, req, act, "parse_sandbox_outputs", err)
		return resp, err
	}

	prReq := &runnerv1.ProvisionRunnerRequest{
		OrgId:         orgID,
		AppId:         appID,
		InstallId:     installID,
		OdrIamRoleArn: tfOutputs.OdrIAMRoleArn,
		ClusterInfo: &runnerv1.KubeClusterInfo{
			Id:             tfOutputs.ClusterID,
			Endpoint:       tfOutputs.ClusterEndpoint,
			CaData:         tfOutputs.ClusterCA,
			TrustedRoleArn: w.cfg.NuonAccessRoleArn,
		},
	}
	if _, err = execProvisionRunner(ctx, w.cfg, prReq); err != nil {
		err = fmt.Errorf("unable to provision install runner: %w", err)
		w.finishWithErr(ctx, req, act, "provision_install_runner", err)
		return resp, err
	}

	finishReq := FinishRequest{
		ProvisionRequest:              req,
		InstallationsBucket:           w.cfg.InstallationsBucket,
		InstallationsAccessIAMRoleARN: fmt.Sprintf(w.cfg.OrgInstallationsRoleTemplate, orgID),
		Success:                       true,
	}
	if _, err = execFinish(ctx, act, finishReq); err != nil {
		l.Debug("unable to execute finish step: %w", err)
		return resp, fmt.Errorf("unable to execute finish activity: %w", err)
	}
	l.Debug("finished provisioning", "response", resp)
	return resp, nil
}

func (w wkflow) finishWithErr(ctx workflow.Context, req *installsv1.ProvisionRequest, act *ProvisionActivities, step string, err error) {
	l := workflow.GetLogger(ctx)
	finishReq := FinishRequest{
		ProvisionRequest:    req,
		InstallationsBucket: w.cfg.InstallationsBucket,
		Success:             false,
		ErrorStep:           step,
		ErrorMessage:        fmt.Sprintf("%s", err),
	}

	if resp, execErr := execFinish(ctx, act, finishReq); execErr != nil {
		l.Debug("unable to finish with error: %w", execErr, resp)
	}
}

// exec start executes the start activity
func execStartWorkflow(ctx workflow.Context, act *ProvisionActivities, req StartWorkflowRequest) (StartWorkflowResponse, error) {
	var resp StartWorkflowResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing start workflow activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.StartWorkflow, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

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

// exec finish executes the finish activity
func execFinish(ctx workflow.Context, act *ProvisionActivities, req FinishRequest) (FinishResponse, error) {
	var resp FinishResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing finish workflow activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.Finish, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
