package awaitrunnerhealthy

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
)

const SignalType signal.SignalType = "await-runner-healthy"

type Signal struct {
	InstallID      string `json:"install_id"`
	WorkflowStepID string `json:"workflow_step_id"`

	v *validator.Validate
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithParams = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)

func (s *Signal) WithParams(params *signal.Params) {
	s.v = params.V
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
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

	// Update the workflow step target if step ID is available
	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   runner.ID,
			StepTargetType: "runners",
		}); err != nil {
			return errors.Wrap(err, "unable to update workflow step target")
		}
	}

	// Poll for runner health
	if err := poll.Poll(ctx, s.v, poll.PollOpts{
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
