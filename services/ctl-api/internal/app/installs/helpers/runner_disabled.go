package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// ErrRunnerDisabled is returned when a workflow that needs the install runner is
// requested while the stack has the runner disabled.
var ErrRunnerDisabled = errors.New("the install runner is disabled")

const runnerDisabledDescription = "This install's runner is disabled, so it cannot run jobs. Re-enable it in the install stack, then try again."

// NewRunnerDisabledConflict wraps ErrRunnerDisabled as a 409 with actionable
// copy, so every surface reports the same reason and remedy.
func NewRunnerDisabledConflict() error {
	return stderr.ErrConflict{
		Err:         ErrRunnerDisabled,
		Description: runnerDisabledDescription,
	}
}

// IsRunnerDisabled reports whether the install's runner was disabled by its
// stack. A missing runner is not disabled — it has simply not been provisioned
// yet, which the provisioning lifecycle already models.
func (s *Helpers) IsRunnerDisabled(ctx context.Context, installID string) (bool, error) {
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

	return statuses[0] == app.RunnerStatusDisabled, nil
}
