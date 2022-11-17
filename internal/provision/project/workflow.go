package project

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-apps/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type ProvisionProjectRequest struct {
	DryRun bool `json:"dry_run"`

	OrgID string `json:"org_id" validate:"required"`
	AppID string `json:"app_id" validate:"required"`
}

func (r ProvisionProjectRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type ProvisionProjectResponse struct{}

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) ProvisionProject(ctx workflow.Context, req ProvisionProjectRequest) (ProvisionProjectResponse, error) {
	resp := ProvisionProjectResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	act := NewActivities()

	cwpRequest := CreateWaypointProjectRequest{
		TokenSecretNamespace: w.cfg.WaypointTokenNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID),
		OrgID:                req.OrgID,
		AppID:                req.AppID,
	}
	cwpResp, err := execCreateWaypointProject(ctx, act, cwpRequest)
	if err != nil {
		return resp, fmt.Errorf("failed to create waypoint project: %w", err)
	}
	l.Debug("successfully created waypoint project: %w", cwpResp)

	uwwRequest := UpsertWaypointWorkspaceRequest{
		TokenSecretNamespace: w.cfg.WaypointTokenNamespace,
		OrgServerAddr:        waypoint.DefaultOrgServerAddress(w.cfg.WaypointServerRootDomain, req.OrgID),
		OrgID:                req.OrgID,
		AppID:                req.AppID,
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
