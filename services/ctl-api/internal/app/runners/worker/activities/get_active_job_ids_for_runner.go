package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetActiveJobIDsForRunner(ctx context.Context, runnerID string) ([]string, error) {
	var jobs []app.RunnerJob
	res := a.db.WithContext(ctx).
		Select("id").
		Where(&app.RunnerJob{RunnerID: runnerID}).
		Where("status IN ?", []string{
			string(app.RunnerJobStatusAvailable),
			string(app.RunnerJobStatusQueued),
			string(app.RunnerJobStatusInProgress),
		}).
		Find(&jobs)
	if res.Error != nil {
		return nil, res.Error
	}

	ids := make([]string, len(jobs))
	for i, j := range jobs {
		ids[i] = j.ID
	}
	return ids, nil
}
