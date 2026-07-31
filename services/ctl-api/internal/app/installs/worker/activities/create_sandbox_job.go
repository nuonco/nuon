package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateSandboxJobRequest struct {
	InstallID   string                     `validate:"required"`
	RunnerID    string                     `validate:"required"`
	OwnerType   string                     `validate:"required"`
	OwnerID     string                     `validate:"required"`
	JobType     app.RunnerJobType          `validate:"omitempty"`
	Op          app.RunnerJobOperationType `validate:"required"`
	Metadata    map[string]string          `validate:"required"`
	LogStreamID string                     `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateSandboxJob(ctx context.Context, req *CreateSandboxJobRequest) (*app.RunnerJob, error) {
	jobType := req.JobType
	if jobType == "" {
		jobType = app.RunnerJobTypeSandboxTerraform
	}

	var run app.InstallSandboxRun
	res := a.db.WithContext(ctx).
		Select("install_id", "install_workflow_id").
		Where(app.InstallSandboxRun{ID: req.OwnerID}).
		First(&run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install sandbox run: %w", res.Error)
	}
	if run.InstallID != req.InstallID {
		return nil, fmt.Errorf("sandbox run install %s does not match request install %s", run.InstallID, req.InstallID)
	}

	if run.InstallWorkflowID != nil {
		ctx = cctx.SetFlowWorkflowIDContext(ctx, *run.InstallWorkflowID)
	}
	ctx = cctx.SetFlowInstallIDContext(ctx, run.InstallID)

	job, err := a.runnersHelpers.CreateInstallSandboxJob(ctx,
		req.RunnerID,
		req.OwnerType,
		req.OwnerID,
		jobType,
		req.Op,
		req.Metadata,
		req.LogStreamID,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
