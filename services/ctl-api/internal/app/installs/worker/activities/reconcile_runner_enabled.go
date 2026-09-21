package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type ReconcileRunnerEnabled struct {
	RunnerID string `json:"runner_id" validate:"required"`
	Disabled bool   `json:"disabled"`
}

// ReconcileRunnerEnabled aligns a runner's status with the runner_enabled value
// from its install stack outputs. Re-enabling returns the runner to pending
// rather than active: pending is skipped by the health machinery the same way
// disabled is, so the runner only becomes active once it actually reports in.
//
// @temporal-gen-v2 activity
// @max-retries 2
// @local
func (a *Activities) ReconcileRunnerEnabled(ctx context.Context, req *ReconcileRunnerEnabled) error {
	var runner app.Runner
	if res := a.db.WithContext(ctx).First(&runner, "id = ?", req.RunnerID); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to get runner")
	}

	target := app.RunnerStatusDisabled
	description := "runner is disabled by the install stack"
	switch {
	case req.Disabled:
		if runner.Status == app.RunnerStatusDisabled {
			return nil
		}
	case runner.Status == app.RunnerStatusDisabled:
		target = app.RunnerStatusPending
		description = "runner was re-enabled by the install stack"
	default:
		return nil
	}

	if _, err := a.statusActivities.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          req.RunnerID,
		Status:            target,
		StatusDescription: description,
		Metadata:          map[string]any{app.RunnerOfflineTSMetadataKey: nil},
	}); err != nil {
		return fmt.Errorf("unable to transition runner status: %w", err)
	}

	return nil
}
