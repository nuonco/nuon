package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/runner/v1"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
)

type RunnerConfig struct {
	ID            string `json:"id" validate:"required"`
	Cookie        string `json:"cookie" validate:"required"`
	ServerAddr    string `json:"server_addr" validate:"required"`
	OdrIAMRoleArn string `json:"odr_iam_role_arn" validate:"required"`
}

func (r RunnerConfig) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
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
//nolint:funlen
func (w wkflow) ProvisionRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
	resp := &runnerv1.ProvisionRunnerResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	// get waypoint server cookie
	l := log.With(workflow.GetLogger(ctx))

	// parse IDs into short IDs, and use them for all subsequent requests
	orgServerAddr := waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId)
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
	act := NewActivities(nil, workers.Config{})

	// get waypoint server cookie
	gwscReq := GetWaypointServerCookieRequest{
		TokenSecretNamespace: w.cfg.WaypointBootstrapTokenNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		ClusterInfo:          clusterInfo,
	}

	gwscResp, err := getWaypointServerCookie(ctx, act, gwscReq)
	if err != nil {
		err = fmt.Errorf("failed to get waypoint server cookie: %w", err)
		l.Debug(err.Error())
		return resp, err
	}
	l.Debug("successfully fetched waypoint server cookie")

	// install waypoint
	wpChart, err := helm.LoadChart(w.cfg.WaypointChartDir)
	chart := &helm.Chart{
		Name:    wpChart.Metadata.Name,
		Version: wpChart.Metadata.Version,
		Dir:     w.cfg.WaypointChartDir,
	}
	iwReq := InstallWaypointRequest{
		Namespace:   req.OrgId,
		ReleaseName: fmt.Sprintf("wp-%s-runner", req.OrgId),
		Chart:       chart,
		Atomic:      false,
		OrgID:       req.OrgId,
		ClusterInfo: clusterInfo,
		RunnerConfig: RunnerConfig{
			Cookie:        gwscResp.Cookie,
			ID:            req.OrgId,
			ServerAddr:    orgServerAddr,
			OdrIAMRoleArn: req.OdrIamRoleArn,
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
		ClusterInfo:          clusterInfo,
		OrgID:                req.OrgId,
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
		OrgID:                req.OrgId,
		ClusterInfo:          clusterInfo,
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
		OrgID:                req.OrgId,
		ClusterInfo:          clusterInfo,
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
		OrgID:                req.OrgId,
		NamespaceName:        req.OrgId,
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
