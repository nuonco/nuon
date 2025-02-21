package jobloop

import (
	"time"

	smithytime "github.com/aws/smithy-go/time"
	"github.com/sourcegraph/conc/panics"
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

			if err := smithytime.SleepWithContext(j.ctx, defaultJobPollBackoff); err != nil {
				return
			}
			continue
		}

		if len(jobs) < 1 {
			if err := smithytime.SleepWithContext(j.ctx, starvedJobPollBackoff); err != nil {
				return
			}
			continue
		}

		job := jobs[0]

		// execute the job
		var pc panics.Catcher
		pc.Try(func() {
			err = j.executeJob(j.ctx, job)
		})
		if err != nil {
			j.errRecorder.Record("job failed", err)
		}

		// if a panic is _recorded_ we do not restart the runner automatically.
		if rc := pc.Recovered(); err != nil {
			j.l.Error("job panic",
				zap.String("stack-trace", rc.String()),
				zap.String("job-type", string(job.Type)),
				zap.String("job-group", string(job.Group)),
			)
		}

		// iterate for the next loop
		if err := smithytime.SleepWithContext(j.ctx, defaultJobPollBackoff); err != nil {
			return
		}

	}
}
