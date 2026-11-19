package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// IsRunnerNotUsed reports whether the install's runner status is not-used:
// the install's jobs run on control-plane install workers instead of a live
// runner process, so checks that require a live runner process (HasActiveRunner,
// the runbook gate, the group directive) don't apply to it. A missing runner
// reports false -- it simply has not been provisioned yet, which the normal
// provisioning lifecycle already models. Disabled runners also report false:
// "disabled = frozen install" is a separate check (IsRunnerDisabled) that a
// not-used runner can never match.
func (s *Helpers) IsRunnerNotUsed(ctx context.Context, installID string) (bool, error) {
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

	return statuses[0] == app.RunnerStatusNotUsed, nil
}
