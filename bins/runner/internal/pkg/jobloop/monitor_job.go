package jobloop

import (
	"context"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"
)

func (j *jobLoop) monitorJob(ctx context.Context, cancel func(), doneCh chan struct{}, jobID string, l *zap.Logger) {
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

		if job.Status == models.AppRunnerJobStatusCancelled {
			l.Error("job was cancelled via API, attempting to cancel execution")
			cancel()
			return
		}

		if job.Status == models.AppRunnerJobStatusTimedDashOut {
			l.Error("job was timed out via API, attempting to cancel execution")
			cancel()
			return
		}

		if job.Status == models.AppRunnerJobStatusFailed {
			l.Error("job was failed via API, attempting to cancel execution")
			cancel()
			return
		}
	}
}
