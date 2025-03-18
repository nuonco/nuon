package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

const (
	// this means that any job more than 30m will be disgarded when showing the queue depth
	discardJobDuration time.Duration = time.Minute * 30
)

type GetRunnerShutdownJobQueueRequest struct {
	RunnerID string `validate:"required"`
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) GetRunnerShutdownJobQueue(ctx context.Context, req *GetRunnerShutdownJobQueueRequest) ([]*app.RunnerJob, error) {
	// Get queued, available, and in progress shutdown jobs from the operation gruop
	var jobs []*app.RunnerJob
	res := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Where(
			app.RunnerJob{
				RunnerID: req.RunnerID,
				Group:    app.RunnerJobGroupOperations,
				Type:     app.RunnerJobTypeShutDown,
			}).Where(
		`created_at > ? AND status IN ?`,
		time.Now().Add(-discardJobDuration).Format(time.RFC3339), []app.RunnerJobStatus{
			app.RunnerJobStatusQueued,
			app.RunnerJobStatusAvailable,
			app.RunnerJobStatusInProgress,
		}).Order("created_at desc").Find(&jobs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job queue")
	}

	return jobs, nil
}
