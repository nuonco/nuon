package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateHealthCheckJobRequest struct {
	RunnerID    string `validate:"required"`
	OwnerID     string `validate:"required"`
	LogStreamID string `validate:"required"`
	Metadata    map[string]string
}

// @temporal-gen activity
func (a *Activities) CreateHealthCheckJob(ctx context.Context, req *CreateHealthCheckJobRequest) (*app.RunnerJob, error) {
	return a.helpers.CreateHealthCheckJob(ctx, req.RunnerID, req.RunnerID, req.LogStreamID, req.Metadata)
}
