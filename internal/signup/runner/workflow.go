package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	waypointhelm "github.com/powertoolsdev/go-helm/waypoint"
	"github.com/powertoolsdev/go-kube"
	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-orgs/internal"
)

type RunnerConfig struct {
	ID         string
	Cookie     string
	ServerAddr string
}

// RunnerRequest includes the set of arguments needed to provision a sandbox
type InstallRunnerRequest struct {
	OrgID string `json:"org_id" validate:"required"`
}

func (i InstallRunnerRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(i)
}

type InstallRunnerResponse struct {
	TerraformOutputs map[string]string
}

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// Runner is a workflow that creates an app install sandbox using terraform
//
//nolint:funlen //NOTE(cp): this will be fixed with child workflows eventually
func (w wkflow) Install(ctx workflow.Context, req InstallRunnerRequest) (InstallRunnerResponse, error) {
	resp := InstallRunnerResponse{}

	// get waypoint server cookie
	l := log.With(workflow.GetLogger(ctx))

	// parse IDs into short IDs, and use them for all subsequent requests
	orgServerAddr := waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID)
	clusterInfo := kube.ClusterInfo{
		ID:             w.cfg.OrgsK8sClusterID,
		Endpoint:       w.cfg.OrgsK8sPublicEndpoint,
		CAData:         w.cfg.OrgsK8sCAData,
		TrustedRoleARN: w.cfg.OrgsK8sRoleArn,
	}

	l.Debug("installing org waypoint runner")
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// NOTE(jdt): this is just so that we can use the method names
	// the actual struct isn't used by temporal during dispatch at all
	act := NewActivities(workers.Config{})

	// get waypoint server cookie
	gwscReq := GetWaypointServerCookieRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgID,
	}

	gwscResp, err := getWaypointServerCookie(ctx, act, gwscReq)
	if err != nil {
		err = fmt.Errorf("failed to get waypoint server cookie: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully fetched waypoint server cookie")

	// install waypoint
	iwReq := InstallWaypointRequest{
		Namespace:   req.OrgID,
		ReleaseName: fmt.Sprintf("wp-%s-runner", req.OrgID),
		Chart:       &waypointhelm.DefaultChart,
		Atomic:      false,
		OrgID:       req.OrgID,
		ClusterInfo: clusterInfo,
		RunnerConfig: RunnerConfig{
			Cookie:     gwscResp.Cookie,
			ID:         req.OrgID,
			ServerAddr: orgServerAddr,
		},
	}
	_, err = installWaypoint(ctx, act, iwReq)
	if err != nil {
		err = fmt.Errorf("failed to install waypoint: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully installed waypoint runner")

	awrReq := AdoptWaypointRunnerRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgID,
	}
	_, err = adoptWaypointRunner(ctx, act, awrReq)
	if err != nil {
		err = fmt.Errorf("failed to adopt waypoint runner: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully adopted waypoint runner")

	cscReq := CreateServerConfigRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgID,
	}
	_, err = createServerConfigActivity(ctx, act, cscReq)
	if err != nil {
		err = fmt.Errorf("failed to create server config: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully created server config")

	crpReq := CreateRunnerProfileRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgID,
	}
	_, err = createRunnerProfileActivity(ctx, act, crpReq)
	if err != nil {
		err = fmt.Errorf("failed to create runner profile: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully created runner profile")

	crbReq := CreateRoleBindingRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgID,
		NamespaceName:        req.OrgID,
		ClusterInfo:          clusterInfo,
	}
	_, err = createRoleBinding(ctx, act, crbReq)
	if err != nil {
		err = fmt.Errorf("failed to create role binding: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully created rolebinding for runner")

	return resp, nil
}

// getWaypointServerCookie executes an activity to get the the waypoint server
func getWaypointServerCookie(
	ctx workflow.Context,
	act *Activities,
	req GetWaypointServerCookieRequest,
) (GetWaypointServerCookieResponse, error) {
	var resp GetWaypointServerCookieResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
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
func installWaypoint(
	ctx workflow.Context,
	act *Activities,
	req InstallWaypointRequest,
) (InstallWaypointResponse, error) {
	var resp InstallWaypointResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
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
func adoptWaypointRunner(
	ctx workflow.Context,
	act *Activities,
	req AdoptWaypointRunnerRequest,
) (AdoptWaypointRunnerResponse, error) {
	var resp AdoptWaypointRunnerResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing adopt waypoint runner activity")
	fut := workflow.ExecuteActivity(ctx, act.AdoptWaypointRunner, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createServerConfig creates the org server config
func createServerConfigActivity(
	ctx workflow.Context,
	act *Activities,
	req CreateServerConfigRequest,
) (CreateServerConfigResponse, error) {
	var resp CreateServerConfigResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create server config activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateServerConfig, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createServerConfig creates the org server config
func createRunnerProfileActivity(
	ctx workflow.Context,
	act *Activities,
	req CreateRunnerProfileRequest,
) (CreateRunnerProfileResponse, error) {
	var resp CreateRunnerProfileResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create server config activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateRunnerProfile, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createRoleBinding: creates the rolebinding in the correct namespace
func createRoleBinding(
	ctx workflow.Context,
	act *Activities,
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
