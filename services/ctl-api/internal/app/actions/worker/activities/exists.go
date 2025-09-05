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
	return loop.CheckExists[*app.ActionWorkflow](ctx, a.db, req.ID)
}
