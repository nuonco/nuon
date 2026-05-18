package removeoldrunnerprocesses

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	orgactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "org-remove-old-runner-processes"

type Signal struct {
	OrgID string `json:"org_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.OrgID == "" {
		return fmt.Errorf("org_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	listOpts := &workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
	}
	listResp, err := orgactivities.AwaitListOldRunnerProcessesForOrgByOrgID(ctx, s.OrgID, listOpts)
	if err != nil {
		return fmt.Errorf("unable to list old runner processes: %w", err)
	}

	// passing caller specific opts as some tcclient operations take take longer than other, this allows configuring
	// retry and timeouts better
	termOpts := &workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    5,
		},
	}

	deleted, failed := 0, 0
	for _, p := range listResp.Processes {
		if err := orgactivities.AwaitTerminateAndDeleteRunnerProcess(ctx, orgactivities.TerminateAndDeleteRunnerProcessRequest{
			ProcessID: p.ID,
			RunnerID:  p.RunnerID,
		}, termOpts); err != nil {
			l.Warn("failed to terminate and delete process after retries",
				"process_id", p.ID,
				"runner_id", p.RunnerID,
				"error", err)
			failed++
			continue
		}
		deleted++
	}

	l.Info("old runner processes removal complete",
		"org_id", s.OrgID,
		"processes_deleted", deleted,
		"processes_failed", failed)
	return nil
}
