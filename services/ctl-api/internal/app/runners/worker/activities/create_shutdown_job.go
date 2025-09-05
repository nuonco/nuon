package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateShutdownJobRequest struct {
	RunnerID    string            `validate:"required"`
	OwnerID     string            `validate:"required"`
	LogStreamID string            `validate:"required"`
	Metadata    map[string]string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateShutdownJob(ctx context.Context, req *CreateShutdownJobRequest) (*app.RunnerJob, error) {
	return a.helpers.CreateShutdownJob(ctx, req.RunnerID, req.RunnerID, req.LogStreamID, req.Metadata)
}

type CreateMngJobRequest struct {
	RunnerID    string            `validate:"required"`
	OwnerID     string            `validate:"required"`
	LogStreamID string            `validate:"required"`
	JobType     app.RunnerJobType `validate:"required"`

	Metadata map[string]string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateMngJob(ctx context.Context, req *CreateMngJobRequest) (*app.RunnerJob, error) {
	return a.helpers.CreateMngJob(ctx, req.RunnerID, req.LogStreamID, req.JobType, req.Metadata)
}
