package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
)

// Deprovision is a workflow that creates an app install sandbox using terraform
//
//nolint:funlen
func (w wkflow) DeprovisionRunner(ctx workflow.Context, req *runnerv1.DeprovisionRunnerRequest) (*runnerv1.DeprovisionRunnerResponse, error) {
	resp := &runnerv1.DeprovisionRunnerResponse{}

	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 60 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	// create waypoint project
	switch req.RunnerType {
	case installsv1.RunnerType_RUNNER_TYPE_AWS_ECS:
		if err := w.uninstallECSRunner(ctx, req); err != nil {
			return resp, fmt.Errorf("unable to uninstall ecs runner: %w", err)
		}
	case installsv1.RunnerType_RUNNER_TYPE_AWS_EKS:
		if err := w.uninstallEKSRunner(ctx, req); err != nil {
			return resp, fmt.Errorf("unable to uninstall eks runner: %w", err)
		}
	default:
		return resp, fmt.Errorf("unsupported runner type")
	}

	// forget waypoint runner
	orgServerAddr := client.DefaultOrgServerAddress(w.cfg.OrgServerRootDomain, req.OrgId)
	dwpReq := DestroyWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	var dwpResp DestroyWaypointProjectResponse
	err := w.execWaypointActivity(ctx, w.act.DestroyWaypointProject, dwpReq, &dwpResp)
	if err != nil {
		err = fmt.Errorf("failed to destroy waypoint project: %w", err)
		return resp, err
	}

	awrReq := ForgetWaypointRunnerRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	var fwrResp ForgetWaypointRunnerResponse
	err = w.execWaypointActivity(ctx, w.act.ForgetWaypointRunner, awrReq, &fwrResp)
	if err != nil {
		err = fmt.Errorf("failed to forget waypoint runner: %w", err)
		return resp, err
	}

	drpReq := DeleteRunnerConfigRequest{
		TokenSecretNamespace: w.cfg.TokenSecretNamespace,
		OrgServerAddr:        orgServerAddr,
		OrgID:                req.OrgId,
		InstallID:            req.InstallId,
		ClusterInfo:          w.clusterInfo,
	}
	var drpResp DeleteRunnerConfigResponse
	err = w.execWaypointActivity(ctx, w.act.DeleteRunnerConfig, drpReq, &drpResp)
	if err != nil {
		err = fmt.Errorf("failed to destroy waypoint project: %w", err)
		return resp, err
	}
	return resp, nil
}
