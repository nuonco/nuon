package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/helm"
	waypointhelm "github.com/powertoolsdev/mono/pkg/helm/waypoint"
	"github.com/powertoolsdev/mono/pkg/kube"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
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
//nolint:funlen
func (w wkflow) ProvisionRunner(ctx workflow.Context, req *runnerv1.ProvisionRunnerRequest) (*runnerv1.ProvisionRunnerResponse, error) {
	resp := &runnerv1.ProvisionRunnerResponse{}
	l := workflow.GetLogger(ctx)

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	orgServerAddr := fmt.Sprintf("%s.%s:9701", req.OrgId, w.cfg.OrgServerRootDomain)

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// NOTE(jdt): this is just so that we can use the method names
	// the actual struct isn't used by temporal during dispatch at all
	act := NewActivities(nil, workers.Config{})

	clusterInfo := kube.ClusterInfo{
		ID:             w.cfg.OrgsK8sClusterID,
		Endpoint:       w.cfg.OrgsK8sPublicEndpoint,
		CAData:         w.cfg.OrgsK8sCAData,
		TrustedRoleARN: w.cfg.OrgsK8sRoleArn,
	}

	// create waypoint project
	cwpReq := CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          clusterInfo,
	}
	_, err := createWaypointProject(ctx, act, cwpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint project: %w", err)
		return resp, err
	}

	// create waypoint workspace
	cwwReq := CreateWaypointWorkspaceRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          clusterInfo,
	}
	_, err = createWaypointWorkspace(ctx, act, cwwReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint workspace: %w", err)
		return resp, err
	}

	// get waypoint server cookie
	gwscReq := GetWaypointServerCookieRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		ClusterInfo:          clusterInfo,
	}
	gwscResp, err := getWaypointServerCookie(ctx, act, gwscReq)
	if err != nil {
		err = fmt.Errorf("failed to get waypoint server cookie: %w", err)
		return resp, err
	}

	// install waypoint
	chart := &helm.Chart{
		Name:    waypointhelm.DefaultChart.Name,
		Version: waypointhelm.DefaultChart.Version,
		Dir:     w.cfg.WaypointChartDir,
	}
	iwReq := InstallWaypointRequest{
		InstallID:       req.InstallId,
		Namespace:       req.InstallId,
		ReleaseName:     fmt.Sprintf("wp-%s", req.InstallId),
		Chart:           chart,
		Atomic:          false,
		CreateNamespace: true,
		ClusterInfo: kube.ClusterInfo{
			ID:             req.ClusterInfo.Id,
			Endpoint:       req.ClusterInfo.Endpoint,
			CAData:         req.ClusterInfo.CaData,
			TrustedRoleARN: req.ClusterInfo.TrustedRoleArn,
		},

		RunnerConfig: RunnerConfig{
			OdrIAMRoleArn: req.OdrIamRoleArn,
			Cookie:        gwscResp.Cookie,
			ID:            req.InstallId,
			ServerAddr:    orgServerAddr,
		},
	}
	_, err = installWaypoint(ctx, act, iwReq)
	if err != nil {
		err = fmt.Errorf("failed to install waypoint: %w", err)
		return resp, err
	}

	awrReq := AdoptWaypointRunnerRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          clusterInfo,
	}
	_, err = adoptWaypointRunner(ctx, act, awrReq)
	if err != nil {
		err = fmt.Errorf("failed to adopt waypoint runner: %w", err)
		return resp, err
	}

	crbReq := CreateRoleBindingRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		InstallID:            req.InstallId,
		NamespaceName:        req.InstallId,
		ClusterInfo: kube.ClusterInfo{
			ID:             req.ClusterInfo.Id,
			Endpoint:       req.ClusterInfo.Endpoint,
			CAData:         req.ClusterInfo.CaData,
			TrustedRoleARN: req.ClusterInfo.TrustedRoleArn,
		},
	}
	_, err = createRoleBinding(ctx, act, crbReq)
	if err != nil {
		err = fmt.Errorf("failed to create role_binding for runner: %w", err)
		return resp, err
	}

	cwrpReq := CreateWaypointRunnerProfileRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		InstallID:            req.InstallId,
		OrgID:                req.OrgId,
		AwsRegion:            req.Region,
		ClusterInfo:          clusterInfo,
	}
	_, err = createWaypointRunnerProfile(ctx, act, cwrpReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint runner profile: %w", err)
		return resp, err
	}

	l.Debug("finished provisioning", "response", resp)
	return resp, nil
}

// createWaypointProject executes an activity to create the waypoint project on the org's server
func createWaypointProject(ctx workflow.Context, act *Activities, req CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
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
func getWaypointServerCookie(ctx workflow.Context, act *Activities, req GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
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
func installWaypoint(ctx workflow.Context, act *Activities, req InstallWaypointRequest) (InstallWaypointResponse, error) {
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
func adoptWaypointRunner(ctx workflow.Context, act *Activities, req AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
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

// createWaypointWorkspace creates a waypoint workspace
func createWaypointWorkspace(ctx workflow.Context, act *Activities, req CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
	var resp CreateWaypointWorkspaceResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint workspace activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateWaypointWorkspace, req)
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

// createWaypointRunnerProfile: creates the runner profile for this install
func createWaypointRunnerProfile(
	ctx workflow.Context,
	act *Activities,
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
