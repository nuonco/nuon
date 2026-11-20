package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// HasUnfinishedInstallWorkflow reports whether the install has any workflow
// that hasn't reached a terminal status yet. Callers use this to reject
// changes that must not race a workflow already in flight.
func (s *Helpers) HasUnfinishedInstallWorkflow(ctx context.Context, installID string) (bool, error) {
	terminalStatuses := []string{
		string(app.StatusSuccess),
		string(app.StatusError),
		string(app.StatusCancelled),
	}

	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.Workflow{}).
		Where(app.Workflow{OwnerID: installID, OwnerType: "installs"}).
		Where("status->>'status' NOT IN ?", terminalStatuses).
		Count(&count)
	if res.Error != nil {
		return false, errors.Wrap(res.Error, "unable to check for unfinished install workflows")
	}
	return count > 0, nil
}
