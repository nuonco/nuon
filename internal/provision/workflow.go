package provision

import (
	"errors"
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/shortid"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/exp/slices"

	"github.com/powertoolsdev/go-helm/waypoint"
	"github.com/powertoolsdev/go-kube"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

// validInstallRegions are the list of regions allowed for an install to be run in
func validInstallRegions() []string {
	return []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2"}
}

// TODO(jdt): why is this AWS specific?
type AccountSettings struct {
	AwsRegion    string `json:"aws_region"`
	AwsAccountID string `json:"aws_account_id"`
	AwsRoleArn   string `json:"aws_role_arn"`
}

// ProvisionRequest includes the set of arguments needed to provision a sandbox
type ProvisionRequest struct {
	OrgID     string `json:"org_id" validate:"required"`
	AppID     string `json:"app_id" validate:"required"`
	InstallID string `json:"install_id" validate:"required"`

	AccountSettings *AccountSettings `json:"account_settings" validate:"required"`

	SandboxSettings struct {
		Name    string `json:"name" validate:"required"`
		Version string `json:"version" validate:"required"`
	} `json:"sandbox_settings" validate:"required"`
}

type ProvisionResponse struct {
	TerraformOutputs map[string]string
}

const (
	clusterIDKey       = "cluster_id"
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
//
//nolint:funlen,gocyclo
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
	orgServerAddr := fmt.Sprintf("%s.%s:9701", orgID, w.cfg.OrgServerRootDomain)

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

	psReq := ProvisionSandboxRequest{
		AppID:               appID,
		OrgID:               orgID,
		InstallID:           installID,
		BackendBucketName:   w.cfg.InstallationStateBucket,
		BackendBucketRegion: w.cfg.InstallationStateBucketRegion,
		AccountSettings:     req.AccountSettings,
		SandboxSettings:     req.SandboxSettings,
		SandboxBucketName:   w.cfg.SandboxBucket,
		NuonAccessRoleArn:   w.cfg.NuonAccessRoleArn,
	}
	psr, err := provisionSandbox(ctx, act, psReq)
	if err != nil {
		err = fmt.Errorf("unable to provision sandbox: %w", err)
		w.finishWithErr(ctx, req, act, "provision_sandbox", err)
		return resp, err
	}
	resp.TerraformOutputs = psr.Outputs

	if err = checkKeys(psr.Outputs, []string{clusterIDKey, clusterEndpointKey, clusterCAKey}); err != nil {
		err = fmt.Errorf("missing necessary TF output to continue: %w", err)
		w.finishWithErr(ctx, req, act, "check_terraform_outputs", err)
		return resp, err
	}

	// create waypoint project
	cwpReq := CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                orgID,
		InstallID:            installID,
	}
	_, err = createWaypointProject(ctx, act, cwpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint project: %w", err)
		w.finishWithErr(ctx, req, act, "create_waypoint_project", err)
		return resp, err
	}

	// create waypoint workspace
	cwwReq := CreateWaypointWorkspaceRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                orgID,
		InstallID:            installID,
	}
	_, err = createWaypointWorkspace(ctx, act, cwwReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint workspace: %w", err)
		w.finishWithErr(ctx, req, act, "create_waypoint_workspace", err)
		return resp, err
	}

	// get waypoint server cookie
	gwscReq := GetWaypointServerCookieRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                orgID,
	}
	gwscResp, err := getWaypointServerCookie(ctx, act, gwscReq)
	if err != nil {
		err = fmt.Errorf("failed to get waypoint server cookie: %w", err)
		w.finishWithErr(ctx, req, act, "get_waypoint_server_cookie", err)
		return resp, err
	}

	// install waypoint
	iwReq := InstallWaypointRequest{
		Namespace:       installID,
		ReleaseName:     fmt.Sprintf("wp-%s", installID),
		Chart:           &waypoint.DefaultChart,
		Atomic:          false,
		CreateNamespace: true,

		ClusterInfo: kube.ClusterInfo{
			ID:       psr.Outputs[clusterIDKey],
			Endpoint: psr.Outputs[clusterEndpointKey],
			CAData:   psr.Outputs[clusterCAKey],
		},

		RunnerConfig: RunnerConfig{
			Cookie:     gwscResp.Cookie,
			ID:         installID,
			ServerAddr: orgServerAddr,
		},
	}
	_, err = installWaypoint(ctx, act, iwReq)
	if err != nil {
		err = fmt.Errorf("failed to install waypoint: %w", err)
		w.finishWithErr(ctx, req, act, "install_waypoint", err)
		return resp, err
	}

	awrReq := AdoptWaypointRunnerRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                orgID,
		InstallID:            installID,
	}
	_, err = adoptWaypointRunner(ctx, act, awrReq)
	if err != nil {
		err = fmt.Errorf("failed to adopt waypoint runner: %w", err)
		w.finishWithErr(ctx, req, act, "adopt_waypoint_runner", err)
		return resp, err
	}

	crbReq := CreateRoleBindingRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		InstallID:            installID,
		NamespaceName:        installID,
		ClusterInfo: kube.ClusterInfo{
			ID:             psr.Outputs[clusterIDKey],
			Endpoint:       psr.Outputs[clusterEndpointKey],
			CAData:         psr.Outputs[clusterCAKey],
			TrustedRoleARN: w.cfg.NuonAccessRoleArn,
		},
	}
	_, err = createRoleBinding(ctx, act, crbReq)
	if err != nil {
		err = fmt.Errorf("failed to create role_binding for runner: %w", err)
		w.finishWithErr(ctx, req, act, "create_role_binding", err)
		return resp, err
	}

	cwrpReq := CreateWaypointRunnerProfileRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		InstallID:            installID,
		OrgID:                orgID,
	}
	_, err = createWaypointRunnerProfile(ctx, act, cwrpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint runner profile: %w", err)
		w.finishWithErr(ctx, req, act, "create_waypoint_runner_profile", err)
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

// provisionSandbox executes a provision sandbox activity
func provisionSandbox(ctx workflow.Context, act *ProvisionActivities, req ProvisionSandboxRequest) (ProvisionSandboxResponse, error) {
	var resp ProvisionSandboxResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing provision sandbox activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ProvisionSandbox, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// createWaypointProject executes an activity to create the waypoint project on the org's server
func createWaypointProject(ctx workflow.Context, act *ProvisionActivities, req CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	l := workflow.GetLogger(ctx)

	if err := validateCreateWaypointProjectRequest(req); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint project activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateWaypointProject, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// getWaypointServerCookie executes an activity to get the the waypoint server
func getWaypointServerCookie(ctx workflow.Context, act *ProvisionActivities, req GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
	var resp GetWaypointServerCookieResponse
	l := workflow.GetLogger(ctx)

	if err := validateGetWaypointServerCookieRequest(req); err != nil {
		return resp, err
	}

	l.Debug("executing get waypoint server cookie")
	fut := workflow.ExecuteActivity(ctx, act.GetWaypointServerCookie, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// installWaypoint executes an activity to install waypoint into the sandbox
func installWaypoint(ctx workflow.Context, act *ProvisionActivities, req InstallWaypointRequest) (InstallWaypointResponse, error) {
	var resp InstallWaypointResponse
	l := workflow.GetLogger(ctx)

	if err := validateInstallWaypointRequest(req); err != nil {
		return resp, err
	}

	l.Debug("executing install waypoint activity")
	fut := workflow.ExecuteActivity(ctx, act.InstallWaypoint, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// adoptWaypointRunner adopts the waypoint runner
func adoptWaypointRunner(ctx workflow.Context, act *ProvisionActivities, req AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
	var resp AdoptWaypointRunnerResponse
	l := workflow.GetLogger(ctx)

	if err := validateAdoptWaypointRunnerRequest(req); err != nil {
		return resp, err
	}

	l.Debug("executing adopt waypoint runner activity")
	fut := workflow.ExecuteActivity(ctx, act.AdoptWaypointRunner, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointWorkspace creates a waypoint workspace
func createWaypointWorkspace(ctx workflow.Context, act *ProvisionActivities, req CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
	var resp CreateWaypointWorkspaceResponse
	l := workflow.GetLogger(ctx)

	if err := validateCreateWaypointWorkspaceRequest(req); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint workspace activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateWaypointWorkspace, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

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

// createRoleBinding: creates the rolebinding in the correct namespace
func createRoleBinding(
	ctx workflow.Context,
	act *ProvisionActivities,
	req CreateRoleBindingRequest,
) (CreateRoleBindingResponse, error) {
	var resp CreateRoleBindingResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create role binding activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateRoleBinding, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointRunnerProfile: creates the runner profile for this install
func createWaypointRunnerProfile(
	ctx workflow.Context,
	act *ProvisionActivities,
	req CreateWaypointRunnerProfileRequest,
) (CreateWaypointRunnerProfileResponse, error) {
	var resp CreateWaypointRunnerProfileResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint runner profile activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateWaypointRunnerProfile, req)
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
