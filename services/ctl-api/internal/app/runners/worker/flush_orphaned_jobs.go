package worker

import (
	"time"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	// any job over 12 hours old that is _still_ queued will be automatically flushed
	orphanedJobsThreshold time.Duration = time.Hour * 12
)

// @temporal-gen workflow
func (w *Workflows) FlushOrphanedJobs(ctx workflow.Context, sreq signals.RequestSignal) error {
	ts := workflow.Now(ctx)
	threshold := ts.Add(-orphanedJobsThreshold)

	if err := activities.AwaitFlushOrphanedJobs(ctx, activities.FlushOrphanedJobsRequest{
		RunnerID:  sreq.ID,
		Threshold: threshold,
	}); err != nil {
		return errors.Wrap(err, "unable to flush orphaned jobs")
	}

	return nil
}
