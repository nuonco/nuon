package provision

import (
	"errors"
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/shortid"
	"github.com/powertoolsdev/go-kube"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/exp/slices"

	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/workers-installs/internal/provision/sandbox"
)

// validInstallRegions are the list of regions allowed for an install to be run in
func validInstallRegions() []string {
	return []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2"}
}

// ProvisionRequest includes the set of arguments needed to provision a sandbox
type ProvisionRequest struct {
	OrgID     string `json:"org_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	InstallID string `json:"install_id" validate:"required"`

	AccountSettings *sandbox.AccountSettings `json:"account_settings" validate:"required"`

	SandboxSettings struct {
		Name    string `json:"name" validate:"required"`
		Version string `json:"version" validate:"required"`
	} `json:"sandbox_settings" validate:"required"`
}

type ProvisionResponse struct {
	TerraformOutputs map[string]string
}

const (
	clusterIDKey       = "cluster_name"
	clusterEndpointKey = "cluster_endpoint"
	clusterCAKey       = "cluster_certificate_authority_data"
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
func (w wkflow) Provision(ctx workflow.Context, req ProvisionRequest) (ProvisionResponse, error) {
	resp := ProvisionResponse{}
	l := workflow.GetLogger(ctx)

	if err := validateProvisionRequest(req); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	// parse IDs into short IDs, and use them for all subsequent requests
	orgID, err := shortid.ParseString(req.OrgID)
	if err != nil {
		return resp, fmt.Errorf("unable to get short org ID: %w", err)
	}
	req.OrgID = orgID
	appID, err := shortid.ParseString(req.AppID)
	if err != nil {
		return resp, fmt.Errorf("unable to get short org ID: %w", err)
	}
	req.AppID = appID
	installID, err := shortid.ParseString(req.InstallID)
	if err != nil {
		return resp, fmt.Errorf("unable to get short install ID: %w", err)
	}
	req.InstallID = installID

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// NOTE(jdt): this is just so that we can use the method names
	// the actual struct isn't used by temporal during dispatch at all
	act := NewProvisionActivities(workers.Config{}, nil)

	sReq := StartWorkflowRequest{
		AppID:               appID,
		OrgID:               orgID,
		InstallID:           installID,
		InstallationsBucket: w.cfg.InstallationStateBucket,
		ProvisionRequest:    req,
	}
	_, err = execStartWorkflow(ctx, act, sReq)
	if err != nil {
		err = fmt.Errorf("failed to execute start workflow activity: %w", err)
		w.finishWithErr(ctx, req, act, "start_workflow", err)
		return resp, err
	}

	psReq := sandbox.ProvisionRequest{
		AppID:           appID,
		OrgID:           orgID,
		InstallID:       installID,
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

	if err = checkKeys(psr.TerraformOutputs, []string{clusterIDKey, clusterEndpointKey, clusterCAKey}); err != nil {
		err = fmt.Errorf("missing necessary TF output to continue: %w", err)
		w.finishWithErr(ctx, req, act, "check_terraform_outputs", err)
		return resp, err
	}

	prReq := runner.ProvisionRequest{
		OrgID:     orgID,
		AppID:     appID,
		InstallID: installID,
		ClusterInfo: kube.ClusterInfo{
			ID:       psr.TerraformOutputs[clusterIDKey],
			Endpoint: psr.TerraformOutputs[clusterEndpointKey],
			CAData:   psr.TerraformOutputs[clusterCAKey],
		},
	}
	if _, err = execProvisionRunner(ctx, w.cfg, prReq); err != nil {
		err = fmt.Errorf("unable to provision install runner: %w", err)
		w.finishWithErr(ctx, req, act, "provision_install_runner", err)
		return resp, err
	}

	finishReq := FinishRequest{
		ProvisionRequest:    req,
		InstallationsBucket: w.cfg.InstallationStateBucket,
		Success:             true,
	}
	if _, err = execFinish(ctx, act, finishReq); err != nil {
		l.Debug("unable to execute finish step: %w", err)
		return resp, fmt.Errorf("unable to execute finish activity: %w", err)
	}
	l.Debug("finished provisioning", "response", resp)
	return resp, nil
}

func (w wkflow) finishWithErr(ctx workflow.Context, req ProvisionRequest, act *ProvisionActivities, step string, err error) {
	l := workflow.GetLogger(ctx)
	finishReq := FinishRequest{
		ProvisionRequest:    req,
		InstallationsBucket: w.cfg.InstallationStateBucket,
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
	iwrr sandbox.ProvisionRequest,
) (sandbox.ProvisionResponse, error) {
	var resp sandbox.ProvisionResponse

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
	iwrr runner.ProvisionRequest,
) (runner.ProvisionResponse, error) {
	var resp runner.ProvisionResponse

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

var (
	ErrInvalidInstall         = errors.New("invalid install")
	ErrInvalidAccountSettings = errors.New("invalid account settings")
	ErrInvalidSandboxSettings = errors.New("invalid sandbox settings")
	ErrInvalidBucketName      = errors.New("invalid bucket name")
	ErrInvalidBucketRegion    = errors.New("invalid bucket region")
)

func validateProvisionRequest(req ProvisionRequest) error {
	if req.InstallID == "" {
		return fmt.Errorf("%w: install ID must be set", ErrInvalidInstall)
	}

	// validate account settings
	if req.AccountSettings == nil {
		return fmt.Errorf("%w: account settings must be specified", ErrInvalidAccountSettings)
	}
	if req.AccountSettings.AwsRegion == "" {
		return fmt.Errorf("%w: account region must be set", ErrInvalidAccountSettings)
	}
	if !slices.Contains(validInstallRegions(), req.AccountSettings.AwsRegion) {
		return fmt.Errorf("%w: account region not supported", ErrInvalidAccountSettings)
	}

	if req.AccountSettings.AwsAccountID == "" {
		return fmt.Errorf("%w: account ID must be set", ErrInvalidAccountSettings)
	}
	if req.AccountSettings.AwsRoleArn == "" {
		return fmt.Errorf("%w: account role arn must be set", ErrInvalidAccountSettings)
	}

	// validate sandbox settings
	if req.SandboxSettings.Name == "" {
		return fmt.Errorf("%w: sandbox name must be set", ErrInvalidSandboxSettings)
	}
	if req.SandboxSettings.Version == "" {
		return fmt.Errorf("%w: sandbox version must be set", ErrInvalidSandboxSettings)
	}

	return nil
}

func checkKeys(m map[string]string, keys []string) error {
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return fmt.Errorf("missing key: %s", k)
		}
	}
	return nil
}
