package orgprocesshealthchecksweep

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org_process_healthcheck_sweep"

// maxSweepPages bounds one sweep run; at the default page size that covers 25k
// processes per org.
const maxSweepPages = 50

// Signal checks all active/offline runner processes in an org in paginated batches.
type Signal struct {
	OrgID string `json:"org_id"`
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithMaxInFlightAge = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }

// MaxInFlightAge matches the sweep cron interval.
func (s *Signal) MaxInFlightAge() time.Duration { return 5 * time.Minute }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	cursor := ""
	var totals activities.BatchProcessHealthchecksResponse
	for page := 0; page < maxSweepPages; page++ {
		resp, err := activities.AwaitBatchProcessHealthchecks(ctx, activities.BatchProcessHealthchecksRequest{
			OrgID:    s.OrgID,
			CursorID: cursor,
		})
		if err != nil {
			return fmt.Errorf("unable to run process healthcheck batch: %w", err)
		}

		totals.Checked += resp.Checked
		totals.Active += resp.Active
		totals.Offline += resp.Offline
		totals.Inactive += resp.Inactive
		totals.Shutdowns += resp.Shutdowns
		totals.NoHeartbeat += resp.NoHeartbeat
		totals.MissingQueue += resp.MissingQueue
		totals.Skipped += resp.Skipped
		totals.Errors += resp.Errors

		if resp.Done {
			l.Info("org process healthcheck sweep complete",
				"org_id", s.OrgID,
				"pages", page+1,
				"checked", totals.Checked,
				"active", totals.Active,
				"offline", totals.Offline,
				"inactive", totals.Inactive,
				"shutdowns", totals.Shutdowns,
				"no_heartbeat", totals.NoHeartbeat,
				"missing_queue", totals.MissingQueue,
				"skipped", totals.Skipped,
				"errors", totals.Errors,
			)
			return nil
		}
		cursor = resp.NextCursorID
	}

	l.Warn("org process healthcheck sweep hit max pages",
		"org_id", s.OrgID, "max_pages", maxSweepPages, "checked", totals.Checked)
	return nil
}
