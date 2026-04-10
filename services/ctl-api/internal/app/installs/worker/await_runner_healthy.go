package worker

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
)

// @temporal-gen-v2 workflow
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

	// Determine the process type to poll based on runner group type
	processType := app.InstallProcessForRunnerGroupType(runner.RunnerGroup.Type)
	if processType == app.RunnerProcessTypeUnknown {
		return errors.Errorf("unsupported runner group type %s for health checking", runner.RunnerGroup.Type)
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   runner.ID,
		StepTargetType: plugins.TableName(w.db, runner),
	}); err != nil {
		return errors.Wrap(err, "unable to update step")
	}

	// Poll for runner process health
	processReq := activities.GetCurrentRunnerProcessRequest{
		RunnerID:    runner.ID,
		ProcessType: processType,
	}
	if err := poll.Poll(ctx, w.v, poll.PollOpts{
		MaxTS:           workflow.Now(ctx).Add(time.Hour),
		InitialInterval: time.Second * 15,
		MaxInterval:     time.Minute * 1,
		BackoffFactor:   1.1,
		Fn: func(ctx workflow.Context) error {
			process, err := activities.AwaitGetCurrentRunnerProcess(ctx, processReq)
			if err != nil {
				return err
			}

			if process.ProcessStatus() != app.RunnerProcessStatusActive {
				return errors.Errorf("runner process is not healthy (status: %s)", process.ProcessStatus())
			}
			return nil
		},
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return errors.Wrap(err, "unable to get workflow logger")
			}

			l.Debug("checking runner process status again in "+dur.String(), zap.Duration("duration", dur))
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "runner process was not healthy")
	}

	return nil
}
