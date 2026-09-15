package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

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

	if err := a.statusActivities.UpdateRunnerStatusV2Metadata(ctx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
		RunnerID: req.RunnerID,
		Metadata: map[string]any{
			app.RunnerOfflineTSMetadataKey:         nil,
			app.RunnerOfflineFromStatusMetadataKey: nil,
		},
	}); err != nil {
		return fmt.Errorf("unable to clear runner offline metadata: %w", err)
	}

	res := a.db.WithContext(ctx).Model(&app.Runner{ID: req.RunnerID}).Updates(app.Runner{
		Status:            target,
		StatusDescription: description,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update runner status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound, fmt.Sprintf("no runner found: %s", req.RunnerID))
	}

	if err := a.statusActivities.UpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          req.RunnerID,
		Status:            target,
		StatusDescription: description,
	}); err != nil {
		return fmt.Errorf("unable to update runner status v2: %w", err)
	}

	return nil
}
