package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateJobRequest struct {
	RunnerID    string                     `validate:"required"`
	OwnerType   string                     `validate:"required"`
	OwnerID     string                     `validate:"required"`
	Op          app.RunnerJobOperationType `validate:"required"`
	Type        app.RunnerJobType          `validate:"required"`
	LogStreamID string                     `validate:"required"`
	Metadata    map[string]string          `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateJob(ctx context.Context, req *CreateJobRequest) (*app.RunnerJob, error) {
	runner, err := a.getRunner(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch runner: %w", err)
	}

	ctx = cctx.SetOrgIDContext(ctx, runner.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, runner.CreatedByID)

	job, err := a.helpers.CreateRunnerJob(ctx,
		req.RunnerID,
		req.OwnerType,
		req.OwnerID,
		req.Type,
		req.Op,
		req.LogStreamID,
		req.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create runner job: %w", err)
	}

	return job, nil
}
