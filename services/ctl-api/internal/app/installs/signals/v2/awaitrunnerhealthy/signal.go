package awaitrunnerhealthy

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
)

const SignalType signal.SignalType = "await-runner-healthy"

type Signal struct {
	InstallID      string `json:"install_id"`
	WorkflowStepID string `json:"workflow_step_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	if s.WorkflowStepID == "" {
		return errors.New("workflow_step_id is required")
	}

	// Validate that the install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get the install
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	// Get the runner
	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	// Update the workflow step target
	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.WorkflowStepID,
		StepTargetID:   runner.ID,
		StepTargetType: plugins.TableName(nil, runner),
	}); err != nil {
		return errors.Wrap(err, "unable to update workflow step target")
	}

	// Poll for runner health
	if err := poll.Poll(ctx, nil, poll.PollOpts{
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
				return errors.New("runner is not healthy")
			}
			return nil
		},
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l, err := log.WorkflowLogger(ctx)
			if err != nil {
				return errors.Wrap(err, "unable to get workflow logger")
			}

			l.Debug("checking runner status again", zap.Duration("next_check_in", dur))
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "runner did not become healthy")
	}

	return nil
}
