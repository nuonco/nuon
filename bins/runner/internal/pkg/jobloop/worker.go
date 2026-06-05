package jobloop

import (
	"context"
	"fmt"
	"os"
	"time"

	smithytime "github.com/aws/smithy-go/time"
	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/conc/panics"
	"go.uber.org/fx"
	"go.uber.org/zap"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	defaultJobPollBackoff time.Duration = time.Second * 1
	starvedJobPollBackoff time.Duration = time.Second * 5

	// Tail (long-poll) tuning. The ctl-api endpoint caps server-side
	// hold at 25s and the runner request gets a small grace on top so
	// the server's empty 200 reaches us before the client cancels.
	tailJobPollWait    time.Duration = 25 * time.Second
	tailJobPollTimeout time.Duration = tailJobPollWait + 5*time.Second
)

func (j *jobLoop) runWorker() {
	fmt.Printf("starting worker: %s\n", j.jobGroup)
	l, _ := zap.NewProduction()
	l = l.With(zap.Any("group", j.jobGroup))

	if err := j.worker(); err != nil {
		l.Warn("job loop stopped due to error", zap.Error(err))
	}

	l.Warn("shutting down runner due to closing job loop")
	os.Exit(155)
	if err := j.shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
		l.Warn("unable to shut down", zap.Error(err))
	}
}

func (j *jobLoop) worker() error {
	// useTail flips to false once the org returns 404 from the tail
	// endpoint so we don't keep re-probing a known-disabled endpoint
	// every iteration; falls back to the legacy poll for the lifetime
	// of this process.
	useTail := j.settings != nil && j.settings.LongPollJobs

	for {
		select {
		case <-j.ctx.Done():
			return nil
		default:
		}

		jobs, err := j.fetchAvailableJobs(useTail)
		if err != nil {
			if errors.Is(err, nuonrunner.ErrTailJobsNotAvailable) {
				j.l.Info("tail jobs endpoint disabled by feature flag, falling back to legacy poll")
				useTail = false
				continue
			}
			j.l.Error("unable to fetch jobs", zap.Error(err))

			if err := smithytime.SleepWithContext(j.ctx, defaultJobPollBackoff); err != nil {
				return err
			}
			continue
		}

		if len(jobs) < 1 {
			// On the tail path the server has already held the
			// request open up to `tailJobPollWait`, so reissue
			// immediately instead of sleeping `starvedJobPollBackoff`
			// and reintroducing the 5s pickup latency we just
			// removed. The legacy GetJobs path keeps the old
			// backoff to avoid hammering the API.
			if !useTail {
				if err := smithytime.SleepWithContext(j.ctx, starvedJobPollBackoff); err != nil {
					return err
				}
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
		if rc := pc.Recovered(); rc != nil {
			j.l.Error("job panic",
				zap.String("stack-trace", rc.String()),
				zap.String("job-type", string(job.Type)),
				zap.String("job-group", string(job.Group)),
			)
		}

		// iterate for the next loop
		if err := smithytime.SleepWithContext(j.ctx, defaultJobPollBackoff); err != nil {
			return err
		}
	}
}

// fetchAvailableJobs branches between the long-poll tail endpoint and the
// legacy poll based on the runtime feature flag. The tail path uses a
// longer per-request timeout because the server intentionally holds it
// open. The legacy path keeps the original 5s ceiling.
func (j *jobLoop) fetchAvailableJobs(useTail bool) ([]*models.AppRunnerJob, error) {
	if useTail {
		tctx, cancel := context.WithTimeoutCause(j.ctx, tailJobPollTimeout, errors.Wrapf(context.DeadlineExceeded, "tail poll for jobs in group %s timed out", j.jobGroup))
		defer cancel()
		return j.apiClient.TailJobs(tctx, j.jobGroup, tailJobPollWait)
	}

	tctx, cancel := context.WithTimeoutCause(j.ctx, 5*time.Second, errors.Wrapf(context.DeadlineExceeded, "polling for jobs in group %s timed out", j.jobGroup))
	defer cancel()
	var lim *int64
	return j.apiClient.GetJobs(tctx, j.jobGroup, j.jobStatus, lim)
}
