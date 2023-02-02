package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-common/shortid"
	appv1 "github.com/powertoolsdev/protos/workflows/generated/types/apps/v1"
	workers "github.com/powertoolsdev/workers-apps/internal"
	"github.com/powertoolsdev/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/workers-apps/internal/provision/repository"
)

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) Provision(ctx workflow.Context, req *appv1.ProvisionRequest) (*appv1.ProvisionResponse, error) {
	resp := appv1.ProvisionResponse{}

	if err := req.Validate(); err != nil {
		//%TODO(cp): add zap logger to workflow
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	orgShortID, err := shortid.ParseString(req.OrgId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orgId to shortID: %w", err)
	}
	appShortID, err := shortid.ParseString(req.AppId)
	if err != nil {
		return nil, fmt.Errorf("failed to parse appID to shortID: %w", err)
	}

	prRequest := repository.ProvisionRepositoryRequest{
		OrgID: orgShortID,
		AppID: appShortID,
	}
	prResp, err := execProvisionRepository(ctx, w.cfg, prRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to provision repository: %w", err)
	}
	l.Debug("successfully provisioned repository: %w", prResp)

	ppReq := project.ProvisionProjectRequest{
		OrgID: orgShortID,
		AppID: appShortID,
	}
	ppResp, err := execProvisionProject(ctx, w.cfg, ppReq)
	if err != nil {
		return nil, fmt.Errorf("failed to provision project: %w", err)
	}
	l.Debug("successfully provisioned project: %w", ppResp)

	l.Debug("finished provisioning app", "response", &resp)
	return &resp, nil
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
