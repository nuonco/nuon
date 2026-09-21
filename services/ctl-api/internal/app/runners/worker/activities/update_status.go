package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type UpdateStatusRequest struct {
	RunnerID          string           `validate:"required"`
	Status            app.RunnerStatus `validate:"required"`
	StatusDescription string           `validate:"required"`
	SkipIfDisabled    bool
	Metadata          map[string]any
}

// @temporal-gen-v2 activity
// @max-retries 2
// @local
func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	if _, err := a.statusActivities.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          req.RunnerID,
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
		SkipIfDisabled:    req.SkipIfDisabled,
		Metadata:          req.Metadata,
	}); err != nil {
		return fmt.Errorf("unable to transition runner status: %w", err)
	}
	return nil
}
