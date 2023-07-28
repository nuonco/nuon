package provision

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	appv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	projectv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1/project/v1"
	repov1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1/repository/v1"
	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision/repository"
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
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := w.startWorkflow(ctx, req)
	if err != nil {
		err = fmt.Errorf("unable to start workflow: %w", err)
		return &resp, err
	}

	prRequest := &repov1.ProvisionRepositoryRequest{
		OrgId: req.OrgId,
		AppId: req.AppId,
	}
	prResp, err := execProvisionRepository(ctx, w.cfg, prRequest)
	if err != nil {
		err = fmt.Errorf("failed to provision repository: %w", err)
		w.finishWorkflow(ctx, req, nil, err)
		return nil, err
	}
	l.Debug("successfully provisioned repository: %w", prResp)
	resp.Repository = prResp

	ppReq := &projectv1.ProvisionProjectRequest{
		OrgId: req.OrgId,
		AppId: req.AppId,
	}
	ppResp, err := execProvisionProject(ctx, w.cfg, ppReq)
	if err != nil {
		w.finishWorkflow(ctx, req, nil, err)
		return nil, fmt.Errorf("failed to provision project: %w", err)
	}
	l.Debug("successfully provisioned project: %w", ppResp)

	l.Debug("finished provisioning app", "response", &resp)
	w.finishWorkflow(ctx, req, &resp, err)
	return &resp, nil
}

func execProvisionRepository(
	ctx workflow.Context,
	cfg workers.Config,
	req *repov1.ProvisionRepositoryRequest,
) (*repov1.ProvisionRepositoryResponse, error) {
	resp := &repov1.ProvisionRepositoryResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing provision repository child workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-repository", req.AppId),
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
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
	req *projectv1.ProvisionProjectRequest,
) (*projectv1.ProvisionProjectResponse, error) {
	resp := &projectv1.ProvisionProjectResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing provision project child workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("%s-provision-project", req.AppId),
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
