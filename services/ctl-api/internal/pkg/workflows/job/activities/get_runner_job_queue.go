package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	// this means that any job more than 6 hours old will be disgarded when showing the queue depth
	discardJobDuration time.Duration = time.Hour * 6
	LimitQueueSize                   = 10
)

type GetRunnerJobQueueRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen activity
// @by-id JobID
func (a *Activities) PkgWorkflowsJobGetRunnerJobQueue(ctx context.Context, req *GetRunnerJobQueueRequest) ([]*app.RunnerJob, error) {
	job, err := a.PkgWorkflowsJobGetJob(ctx, &GetJobRequest{
		ID: req.JobID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner job")
	}

	minJobCreatedAt := job.CreatedAt.Add(-discardJobDuration)
	var jobs []*app.RunnerJob
	res := a.db.WithContext(ctx).
		Where("runner_id = ? AND created_at < ? AND created_at > ? AND status IN ?", job.RunnerID, job.CreatedAt, minJobCreatedAt, []app.RunnerJobStatus{
			app.RunnerJobStatusQueued,
			app.RunnerJobStatusAvailable,
			app.RunnerJobStatusInProgress,
		}).
		Where(app.RunnerJob{
			Group: job.Group,
		}).
		Limit(LimitQueueSize).
		Order("created_at desc").Find(&jobs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job queue")
	}

	return jobs, nil
}
