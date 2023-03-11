package repository

import (
	"fmt"
	"time"

	repov1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1/repository/v1"
	workers "github.com/powertoolsdev/mono/services/workers-apps/internal"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

func (w wkflow) ProvisionRepository(ctx workflow.Context, req *repov1.ProvisionRepositoryRequest) (*repov1.ProvisionRepositoryResponse, error) {
	resp := &repov1.ProvisionRepositoryResponse{}

	l := log.With(workflow.GetLogger(ctx))
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	act := NewActivities()

	l.Debug("creating ecr repository")
	crReq := CreateRepositoryRequest{
		OrgID:                req.OrgId,
		AppID:                req.AppId,
		OrgsEcrAccessRoleArn: w.cfg.OrgsEcrAccessRoleArn,
	}
	ecrResp, err := execCreateRepository(ctx, act, crReq)
	if err != nil {
		return resp, fmt.Errorf("failed to create repository: %w", err)
	}
	resp.RegistryId = ecrResp.RegistryID
	resp.RepositoryArn = ecrResp.RepositoryArn
	resp.RepositoryName = ecrResp.RepositoryName
	resp.RepositoryUri = ecrResp.RepositoryURI

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
