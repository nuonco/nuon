package project

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/go-waypoint"
	projectv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1/project/v1"
	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) ProvisionProject(ctx workflow.Context, req *projectv1.ProvisionProjectRequest) (*projectv1.ProvisionProjectResponse, error) {
	resp := &projectv1.ProvisionProjectResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	act := NewActivities()

	pwsRequest := PingWaypointServerRequest{
		Timeout: time.Minute * 15,
		Addr:    waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId),
	}
	pwsResp, err := execPingWaypointServer(ctx, act, pwsRequest)
	if err != nil {
		return resp, fmt.Errorf("failed to ping waypoint server: %w", err)
	}
	l.Debug("successfully pinged waypoint server: %v", pwsResp)

	cwpRequest := CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.WaypointTokenNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId),
		OrgID:                req.OrgId,
		AppID:                req.AppId,
	}
	cwpResp, err := execCreateWaypointProject(ctx, act, cwpRequest)
	if err != nil {
		return resp, fmt.Errorf("failed to create waypoint project: %w", err)
	}
	l.Debug("successfully created waypoint project: %w", cwpResp)

	uwwRequest := UpsertWaypointWorkspaceRequest{
		TokenSecretNamespace: w.cfg.WaypointTokenNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgId),
		OrgID:                req.OrgId,
		AppID:                req.AppId,
	}
	uwwResp, err := execUpsertWaypointWorkspace(ctx, act, uwwRequest)
	if err != nil {
		return resp, fmt.Errorf("failed to upsert waypoint workspace: %w", err)
	}
	l.Debug("successfully upserted waypoint workspace: %w", uwwResp)

	l.Debug("finished provisioning app", "response", resp)
	return resp, nil
}

func execCreateWaypointProject(
	ctx workflow.Context,
	act *Activities,
	req CreateWaypointProjectRequest,
) (CreateWaypointProjectResponse, error) {
	var resp CreateWaypointProjectResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing create waypoint project activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateWaypointProject, req)

	if err := fut.Get(ctx, &resp); err != nil {
		l.Error("error executing do: %s", err)
		return resp, err
	}

	return resp, nil
}

func execUpsertWaypointWorkspace(
	ctx workflow.Context,
	act *Activities,
	req UpsertWaypointWorkspaceRequest,
) (UpsertWaypointWorkspaceResponse, error) {
	var resp UpsertWaypointWorkspaceResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing upsert uwaypoint workspace activity")
	fut := workflow.ExecuteActivity(ctx, act.UpsertWaypointWorkspace, req)

	if err := fut.Get(ctx, &resp); err != nil {
		l.Error("error executing do: %s", err)
		return resp, err
	}

	return resp, nil
}

func execPingWaypointServer(
	ctx workflow.Context,
	act *Activities,
	req PingWaypointServerRequest,
) (PingWaypointServerResponse, error) {
	var resp PingWaypointServerResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing ping waypoint server activity")
	fut := workflow.ExecuteActivity(ctx, act.PingWaypointServer, req)

	if err := fut.Get(ctx, &resp); err != nil {
		l.Error("error executing do: %s", err)
		return resp, err
	}

	return resp, nil
}
