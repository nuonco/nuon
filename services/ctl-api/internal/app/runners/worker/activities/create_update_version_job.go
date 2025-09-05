package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(sdboyer) - version to update to is determined by a call from the runner
// to the API when processing this job, so version is not a parameter here

type CreateUpdateVersionJobRequest struct {
	RunnerID    string            `validate:"required"`
	OwnerID     string            `validate:"required"`
	LogStreamID string            `validate:"required"`
	Metadata    map[string]string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateUpdateVersionJob(ctx context.Context, req *CreateUpdateVersionJobRequest) (*app.RunnerJob, error) {
	return a.CreateJob(ctx, &CreateJobRequest{
		RunnerID:    req.RunnerID,
		OwnerType:   "runners",
		OwnerID:     req.OwnerID,
		Op:          app.RunnerJobOperationTypeExec,
		Type:        app.RunnerJobTypeUpdateVersion,
		LogStreamID: req.LogStreamID,
		Metadata:    req.Metadata,
	})
}
