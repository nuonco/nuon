package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

type CheckExistsRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 1m
// @start-to-close-timeout 10s
// @by-id ID
func (a *Activities) CheckExists(ctx context.Context, req CheckExistsRequest) (bool, error) {
	return loop.CheckExists[*app.App](ctx, a.db, req.ID)
	// var count int64

	// a.db.WithContext(ctx).
	// 	Model(&app.App{}).
	// 	Where("id = ?", req.ID).
	// 	Count(&count)

	// return count > 0, nil
}
