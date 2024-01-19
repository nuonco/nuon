package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(v *validator.Validate, cfg workers.Config) wkflow {
	return wkflow{
		v:   v,
		cfg: cfg,
		act: NewActivities(nil, workers.Config{}),
		clusterInfo: kube.ClusterInfo{
			ID:             cfg.OrgsK8sClusterID,
			Endpoint:       cfg.OrgsK8sPublicEndpoint,
			CAData:         cfg.OrgsK8sCAData,
			TrustedRoleARN: cfg.OrgsK8sRoleArn,
		},
	}
}

type wkflow struct {
	v           *validator.Validate
	cfg         workers.Config
	act         *Activities
	clusterInfo kube.ClusterInfo
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

	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId)
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// create waypoint project
	cwpReq := CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	_, err := w.createWaypointProject(ctx, cwpReq)
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
		ClusterInfo:          w.clusterInfo,
	}
	_, err = w.createWaypointWorkspace(ctx, cwwReq)
	if err != nil {
		err = fmt.Errorf("failed to create waypoint workspace: %w", err)
		return resp, err
	}

	switch req.RunnerType {
	case installsv1.RunnerType_RUNNER_TYPE_AWS_ECS:
		if err := w.installECSRunner(ctx, req); err != nil {
			return resp, fmt.Errorf("unable to install ecs runner: %w", err)
		}
	case installsv1.RunnerType_RUNNER_TYPE_AWS_EKS:
		if err := w.installEKSRunner(ctx, req); err != nil {
			return resp, fmt.Errorf("unable to install eks runner: %w", err)
		}
	default:
		return resp, fmt.Errorf("unsupported runner type")
	}

	l.Debug("finished provisioning", "response", resp)
	return resp, nil
}

// createWaypointProject executes an activity to create the waypoint project on the org's server
func (w *wkflow) createWaypointProject(ctx workflow.Context, req CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint project activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateWaypointProject, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// getWaypointServerCookie executes an activity to get the the waypoint server
func (w *wkflow) getWaypointServerCookie(ctx workflow.Context, req GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
	var resp GetWaypointServerCookieResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing get waypoint server cookie")
	fut := workflow.ExecuteActivity(ctx, w.act.GetWaypointServerCookie, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// installWaypoint executes an activity to install waypoint into the sandbox
func (w *wkflow) installWaypoint(ctx workflow.Context, req InstallWaypointRequest) (InstallWaypointResponse, error) {
	var resp InstallWaypointResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing install waypoint activity")
	fut := workflow.ExecuteActivity(ctx, w.act.InstallWaypoint, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// adoptWaypointRunner adopts the waypoint runner
func (w *wkflow) adoptWaypointRunner(ctx workflow.Context, req AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
	var resp AdoptWaypointRunnerResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing adopt waypoint runner activity")
	fut := workflow.ExecuteActivity(ctx, w.act.AdoptWaypointRunner, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointWorkspace creates a waypoint workspace
func (w *wkflow) createWaypointWorkspace(ctx workflow.Context, req CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
	var resp CreateWaypointWorkspaceResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint workspace activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateWaypointWorkspace, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createRoleBinding: creates the rolebinding in the correct namespace
func (w *wkflow) createRoleBinding(
	ctx workflow.Context,
	req CreateRoleBindingRequest,
) (CreateRoleBindingResponse, error) {
	var resp CreateRoleBindingResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create role binding activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateRoleBinding, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// createWaypointRunnerProfile: creates the runner profile for this install
func (w *wkflow) createWaypointRunnerProfile(
	ctx workflow.Context,
	req CreateWaypointRunnerProfileRequest,
) (CreateWaypointRunnerProfileResponse, error) {
	var resp CreateWaypointRunnerProfileResponse
	l := workflow.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	l.Debug("executing create waypoint runner profile activity")
	fut := workflow.ExecuteActivity(ctx, w.act.CreateWaypointRunnerProfile, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
