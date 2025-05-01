package worker

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/poll"
)

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 30s
func (w *Workflows) AwaitRunnerHealthy(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "no install found")
	}

	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   runner.ID,
		StepTargetType: plugins.TableName(w.db, runner),
	}); err != nil {
		return errors.Wrap(err, "unable to update step")
	}

	if err := poll.Poll(ctx, w.v, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 1,
		BackoffFactor:   1.1,
		Fn: func(ctx workflow.Context) error {
			runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
			if err != nil {
				return err
			}

			if runner.Status != app.RunnerStatusActive {
				return errors.New("runner was not healthy")
			}
			return nil
		},
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return errors.Wrap(err, "unable to get workflow logger")
			}

			l.Debug("checking runner status again in "+dur.String(), zap.Duration("duration", dur))
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "runner was not healthy")
	}

	return nil
}
