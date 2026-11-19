package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type ReconcileUseRunner struct {
	RunnerID string `json:"runner_id" validate:"required"`
	// UseRunner is the use_runner value reported by the install stack outputs.
	// Missing is treated as true by the caller.
	UseRunner bool `json:"use_runner"`
}

// ReconcileUseRunner aligns a runner's status with the use_runner value from
// its install stack outputs. Unlike ReconcileRunnerEnabled, moving to
// not-used is not gated on the runner's current status: a runnerless install
// never goes through the disabled runner_enabled path, so any status except
// disabled can transition straight to not-used. Moving back to a runner only
// applies from not-used, so it can't undo a disabled/re-enabling transition
// handled by ReconcileRunnerEnabled.
//
// @temporal-gen-v2 activity
// @max-retries 2
// @local
func (a *Activities) ReconcileUseRunner(ctx context.Context, req *ReconcileUseRunner) error {
	if !req.UseRunner {
		if _, err := a.statusActivities.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
			RunnerID:          req.RunnerID,
			Status:            app.RunnerStatusNotUsed,
			StatusDescription: "runner is not used by the install stack",
			Metadata:          map[string]any{app.RunnerOfflineTSMetadataKey: nil},
		}); err != nil {
			return fmt.Errorf("unable to transition runner status: %w", err)
		}
		return nil
	}

	notUsed := app.RunnerStatusNotUsed
	if _, err := a.statusActivities.TransitionRunnerStatus(ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          req.RunnerID,
		Status:            app.RunnerStatusAwaitingHeartbeat,
		StatusDescription: "runner was re-enabled by the install stack",
		Metadata:          map[string]any{app.RunnerOfflineTSMetadataKey: nil},
		OnlyIfStatus:      &notUsed,
	}); err != nil {
		return fmt.Errorf("unable to transition runner status: %w", err)
	}

	return nil
}
