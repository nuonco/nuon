package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// ErrRunnerSwitchInProgress is returned when a workflow is requested for an
// install whose desired use_runner setting doesn't match its runner's actual
// status yet -- the switch between runner and runnerless modes (driven by a
// reprovision_stack workflow) hasn't finished reconciling.
var ErrRunnerSwitchInProgress = errors.New("switching between runner and runnerless is in progress")

const runnerSwitchInProgressDescription = "This install is switching between runner and runnerless mode. Wait for the in-progress reprovision to finish, then try again."

// NewRunnerSwitchInProgressConflict wraps ErrRunnerSwitchInProgress as a 409
// with actionable copy.
func NewRunnerSwitchInProgressConflict() error {
	return stderr.ErrConflict{
		Err:         ErrRunnerSwitchInProgress,
		Description: runnerSwitchInProgressDescription,
	}
}

// IsRunnerSwitchInProgress reports whether the install's desired use_runner
// setting (InstallConfig.UsesRunner) disagrees with its runner's actual
// status (RunnerStatusNotUsed vs. anything else). A missing runner (not yet
// provisioned) never counts as a mismatch -- there is nothing to reconcile
// against yet.
func (s *Helpers) IsRunnerSwitchInProgress(ctx context.Context, installID string) (bool, error) {
	groupIDs := s.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: installID, OwnerType: "installs"})

	var statuses []app.RunnerStatus
	res := s.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Where("runner_group_id IN (?)", groupIDs).
		Order("created_at DESC").
		Limit(1).
		Pluck("status", &statuses)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to get install runner status")
	}
	if len(statuses) == 0 {
		return false, nil
	}

	cfg, err := s.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return false, err
	}

	desiredNotUsed := !cfg.UsesRunner()
	actualNotUsed := statuses[0] == app.RunnerStatusNotUsed
	return desiredNotUsed != actualNotUsed, nil
}
