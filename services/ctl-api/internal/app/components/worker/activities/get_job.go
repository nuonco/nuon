package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetJob(ctx context.Context, req *GetJobRequest) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := a.db.WithContext(ctx).
		First(&job, "id = ?", req.ID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &job, nil
}
