package project

import (
	"fmt"
	"time"

	workers "github.com/powertoolsdev/workers-apps/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type ProvisionProjectRequest struct {
	DryRun bool `json:"dry_run"`

	OrgID string `json:"org_id" validate:"required"`
	AppID string `json:"app_id" validate:"required"`
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

	cwpRequest := CreateWaypointProjectRequest{}
	cwpResp, err := execCreateWaypointProject(ctx, act, cwpRequest)
	if err != nil {
		return resp, fmt.Errorf("failed to create waypoint project: %w", err)
	}
	l.Debug("successfully created waypoint project: %w", cwpResp)

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
