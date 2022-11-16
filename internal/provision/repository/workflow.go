package repository

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	workers "github.com/powertoolsdev/workers-apps/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type ProvisionRepositoryRequest struct {
	DryRun bool `json:"dry_run"`

	OrgID string `json:"org_id" validate:"required"`
	AppID string `json:"app_id" validate:"required"`
}

func (r ProvisionRepositoryRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type ProvisionRepositoryResponse struct{}

type Workflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) Workflow {
	return Workflow{
		cfg: cfg,
	}
}

func (w Workflow) ProvisionRepository(ctx workflow.Context, req ProvisionRepositoryRequest) (ProvisionRepositoryResponse, error) {
	resp := ProvisionRepositoryResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	act := NewActivities()

	l.Debug("creating ecr repository")
	crReq := CreateRepositoryRequest{
		OrgID:          req.OrgID,
		AppID:          req.AppID,
		OrgsIamRoleArn: w.cfg.OrgsIamRoleArn,
	}
	_, err := execCreateRepository(ctx, act, crReq)
	if err != nil {
		return resp, fmt.Errorf("failed to create repository: %w", err)
	}

	return resp, nil
}

func execCreateRepository(
	ctx workflow.Context,
	act *Activities,
	req CreateRepositoryRequest,
) (CreateRepositoryResponse, error) {
	var resp CreateRepositoryResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing create repository activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateRepository, req)

	if err := fut.Get(ctx, &resp); err != nil {
		l.Error("error executing do: %s", err)
		return resp, err
	}

	return resp, nil
}
