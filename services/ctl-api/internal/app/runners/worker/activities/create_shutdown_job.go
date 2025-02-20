package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateShutdownJobRequest struct {
	RunnerID    string
	OwnerID     string
	LogStreamID string
	Metadata    map[string]string
}

// @temporal-gen activity
func (a *Activities) CreateShutdownJob(ctx context.Context, req *CreateShutdownJobRequest) (*app.RunnerJob, error) {
	return a.helpers.CreateShutdownJob(ctx, req.RunnerID, req.RunnerID, req.LogStreamID, req.Metadata)
}
