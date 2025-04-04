package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) ForceSandboxMode(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	org, err := activities.AwaitGetByOrgID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get org")
	}

	if org.SandboxMode {
		l.Debug("skipping, org is in sandbox mode", zap.String("name", org.Name))
		return nil
	}

	if err := activities.AwaitForceSandboxModeByOrgID(ctx, sreq.ID); err != nil {
		return errors.Wrap(err, "unable to force sandbox mode for org")
	}

	if err := activities.AwaitForceRunnersSandboxModeByOrgID(ctx, sreq.ID); err != nil {
		return errors.Wrap(err, "unable to force sandbox mode for runners for org")
	}

	return nil
}
