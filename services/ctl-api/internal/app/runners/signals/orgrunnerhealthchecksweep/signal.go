package orgrunnerhealthchecksweep

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org_runner_healthcheck_sweep"

// maxSweepPages bounds one sweep run; at the default page size that covers 10k
// runners per org.
const maxSweepPages = 50

// Signal checks all live runners in an org in paginated batches.
type Signal struct {
	OrgID string `json:"org_id"`
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithMaxInFlightAge = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }

// MaxInFlightAge matches the sweep cron interval.
func (s *Signal) MaxInFlightAge() time.Duration { return 15 * time.Minute }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	cursor := ""
	var totals activities.BatchRunnerHealthchecksResponse
	for page := 0; page < maxSweepPages; page++ {
		resp, err := activities.AwaitBatchRunnerHealthchecks(ctx, activities.BatchRunnerHealthchecksRequest{
			OrgID:    s.OrgID,
			CursorID: cursor,
		})
		if err != nil {
			return fmt.Errorf("unable to run runner healthcheck batch: %w", err)
		}

		totals.Checked += resp.Checked
		totals.Healthy += resp.Healthy
		totals.Unhealthy += resp.Unhealthy
		totals.Skipped += resp.Skipped
		totals.AlertsEnqueued += resp.AlertsEnqueued
		totals.AlertsDeduped += resp.AlertsDeduped
		totals.Errors += resp.Errors

		if resp.Done {
			l.Info("org runner healthcheck sweep complete",
				"org_id", s.OrgID,
				"pages", page+1,
				"checked", totals.Checked,
				"healthy", totals.Healthy,
				"unhealthy", totals.Unhealthy,
				"skipped", totals.Skipped,
				"alerts_enqueued", totals.AlertsEnqueued,
				"alerts_deduped", totals.AlertsDeduped,
				"errors", totals.Errors,
			)
			return nil
		}
		cursor = resp.NextCursorID
	}

	l.Warn("org runner healthcheck sweep hit max pages",
		"org_id", s.OrgID, "max_pages", maxSweepPages, "checked", totals.Checked)
	return nil
}
