package jobloop

import (
	"context"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"go.uber.org/zap"
)

func (j *jobLoop) monitorJob(ctx context.Context, cancel func(), doneCh chan struct{}, jobID string, l *zap.Logger, jh jobs.JobHandler) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-doneCh:
			return
		case <-ticker.C:
		}

		job, err := j.apiClient.GetJob(ctx, jobID)
		if err != nil {
			l.Warn("unable to fetch job cancellation status", zap.Error(err))
			continue
		}

		if j.handleJobStatus(ctx, job, jh, cancel, l) {
			return
		}
	}
}

func (j *jobLoop) handleJobStatus(ctx context.Context, job *models.AppRunnerJob, jh jobs.JobHandler, cancel func(), l *zap.Logger) bool {
	switch job.Status {
	case models.AppRunnerJobStatusCancelled:
		l.Error("job was cancelled via API, attempting to cancel execution")
	case models.AppRunnerJobStatusTimedDashOut:
		l.Error("job was timed out via API, attempting to cancel execution")
	case models.AppRunnerJobStatusFailed:
		l.Error("job was failed via API, attempting to cancel execution")
	default:
		return false
	}

	l.Info("attempting graceful shutdown")
	err := jh.GracefulShutdown(ctx, job, l)
	if err != nil {
		l.Error("unable to gracefully shutdown job", zap.Error(err))
	}
	cancel()
	return true
}
