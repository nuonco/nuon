package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

var ErrNoActiveRunner = errors.New("the install runner is not running")

const noActiveRunnerDescription = "This install's runner is not currently active. Wait for the runner to come back online or check the runner status, then try again."

func NewNoActiveRunnerConflict() error {
	return stderr.ErrConflict{
		Err:         ErrNoActiveRunner,
		Description: noActiveRunnerDescription,
	}
}

func (s *Helpers) HasActiveRunner(ctx context.Context, installID string) (bool, error) {
	groupIDs := s.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Select("id").
		Where(app.RunnerGroup{OwnerID: installID, OwnerType: "installs"})

	runnerIDs := s.db.WithContext(ctx).
		Model(&app.Runner{}).
		Scopes(scopes.WithDisableViews).
		Select("id").
		Where("runner_group_id IN (?)", groupIDs)

	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Where("runner_id IN (?)", runnerIDs).
		Where("composite_status->>'status' IN ?", app.ActiveRunnerProcessStatuses()).
		Count(&count)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to check active runner process")
	}
	return count > 0, nil
}
