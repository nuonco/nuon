package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRunnerJobQueueRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen activity
// @by-id JobID
func (a *Activities) GetRunnerJobQueue(ctx context.Context, req *GetRunnerJobQueueRequest) ([]*app.RunnerJob, error) {
	job, err := a.GetJob(ctx, &GetJobRequest{
		ID: req.JobID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner job")
	}

	var jobs []*app.RunnerJob
	res := a.db.WithContext(ctx).Where("runner_id = ? AND created_at < ? AND status IN ?", job.RunnerID, job.CreatedAt, []app.RunnerJobStatus{
		app.RunnerJobStatusQueued,
		app.RunnerJobStatusInProgress,
	}).Order("created_at desc").Find(&jobs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job queue")
	}

	return jobs, nil
}
