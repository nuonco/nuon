package workflows

import (
	"context"

	"github.com/pkg/errors"
	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
)

// EnsureSweepSchedule idempotently creates a Temporal Schedule that fires a health
// sweep workflow on a cron cadence in the given namespace. Safe to call on every
// worker boot - an already-registered schedule is treated as success. The sweep
// workflow itself no-ops when native scheduling is disabled, so the schedule may
// exist regardless of the flag (keeping the flip instant).
func EnsureSweepSchedule(ctx context.Context, tc temporalclient.Client, namespace, scheduleID, cron, workflowName string, arg any) error {
	sc, err := tc.ScheduleClientInNamespace(namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get schedule client")
	}

	_, err = sc.Create(ctx, tclient.ScheduleOptions{
		ID:   scheduleID,
		Spec: tclient.ScheduleSpec{CronExpressions: []string{cron}},
		Action: &tclient.ScheduleWorkflowAction{
			ID:        scheduleID + "-wf",
			Workflow:  workflowName,
			TaskQueue: pkgworkflows.APITaskQueue,
			Args:      []any{arg},
		},
		Overlap: enumsv1.SCHEDULE_OVERLAP_POLICY_SKIP,
	})
	if err != nil && !errors.Is(err, temporal.ErrScheduleAlreadyRunning) {
		return errors.Wrap(err, "unable to create sweep schedule")
	}

	return nil
}
