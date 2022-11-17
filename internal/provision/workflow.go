package provision

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/go-common/shortid"
	workers "github.com/powertoolsdev/workers-apps/internal"
	"github.com/powertoolsdev/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/workers-apps/internal/provision/repository"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type ProvisionRequest struct {
	DryRun bool `json:"dry_run"`

	OrgID string `json:"org_id" validate:"required"`
	AppID string `json:"app_id" validate:"required"`
}

type ProvisionResponse struct{}

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) Provision(ctx workflow.Context, req ProvisionRequest) (ProvisionResponse, error) {
	resp := ProvisionResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	orgShortID, err := shortid.ParseString(req.OrgID)
	if err != nil {
		return resp, fmt.Errorf("failed to parse orgID to shortID: %w", err)
	}
	appShortID, err := shortid.ParseString(req.AppID)
	if err != nil {
		return resp, fmt.Errorf("failed to parse appID to shortID: %w", err)
	}

	prRequest := repository.ProvisionRepositoryRequest{
		OrgID: orgShortID,
		AppID: appShortID,
	}
	prResp, err := execProvisionRepository(ctx, w.cfg, prRequest)
	if err != nil {
		return resp, fmt.Errorf("failed to provision repository: %w", err)
	}
	l.Debug("successfully provisioned repository: %w", prResp)

	ppReq := project.ProvisionProjectRequest{
		OrgID: orgShortID,
		AppID: appShortID,
	}
	ppResp, err := execProvisionProject(ctx, w.cfg, ppReq)
	if err != nil {
		return resp, fmt.Errorf("failed to provision project: %w", err)
	}
	l.Debug("successfully provisioned project: %w", ppResp)

	l.Debug("finished provisioning app", "response", resp)
	return resp, nil
}

func execProvisionRepository(
	ctx workflow.Context,
	cfg workers.Config,
	req repository.ProvisionRepositoryRequest,
) (repository.ProvisionRepositoryResponse, error) {
	var resp repository.ProvisionRepositoryResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing provision repository child workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := repository.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionRepository, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execProvisionProject(
	ctx workflow.Context,
	cfg workers.Config,
	req project.ProvisionProjectRequest,
) (project.ProvisionProjectResponse, error) {
	var resp project.ProvisionProjectResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing provision project child workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	wkflow := project.NewWorkflow(cfg)
	fut := workflow.ExecuteChildWorkflow(ctx, wkflow.ProvisionProject, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
