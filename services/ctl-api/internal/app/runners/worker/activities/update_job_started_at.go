package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateJobStartedAtRequest struct {
	JobID     string    `validate:"required"`
	StartedAt time.Time `validate:"required"`
}

// @temporal-gen activity
// @by-id JobID
func (a *Activities) UpdateJobStartedAt(ctx context.Context, req UpdateJobStartedAtRequest) error {
	runner := app.RunnerJob{
		ID: req.JobID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.RunnerJob{
		StartedAt: req.StartedAt,
	})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update job started_at")
	}
	if res.RowsAffected < 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, fmt.Sprintf("no job found with id: %s", req.JobID))
	}

	return nil
}
