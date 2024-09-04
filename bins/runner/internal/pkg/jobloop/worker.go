package jobloop

import (
	"time"

	smithytime "github.com/aws/smithy-go/time"
	"go.uber.org/zap"
)

const (
	defaultJobPollBackoff time.Duration = time.Second * 1
	starvedJobPollBackoff time.Duration = time.Second * 5
)

func (j *jobLoop) worker() {
	for {
		select {
		case <-j.ctx.Done():
			return
		default:
		}

		var lim *int64
		jobs, err := j.apiClient.GetJobs(j.ctx,
			j.jobGroup,
			j.jobStatus,
			lim)
		if err != nil {
			j.l.Error("unable to fetch jobs", zap.Error(err))
			smithytime.SleepWithContext(j.ctx, defaultJobPollBackoff)
			continue
		}

		if len(jobs) < 1 {
			smithytime.SleepWithContext(j.ctx, starvedJobPollBackoff)
			continue
		}

		job := jobs[0]

		if err := j.executeJob(j.ctx, job); err != nil {
			j.errRecorder.Record("job failed", err)
		}
		smithytime.SleepWithContext(j.ctx, defaultJobPollBackoff)
	}
}
