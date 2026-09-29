package awaitrunnerhealthy

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "await-runner-healthy"

// why: Existing await-runner-healthy histories already scheduled their poll
// activities, so the disabled-runner short circuit must not apply on replay.
const skipDisabledRunnerVersion = "await-runner-healthy-skip-disabled-runner-v1"

// why: Offline or error runners will never become healthy during the poll window.
// Old histories that already started the poll loop must not be interrupted.
const failFastUnhealthyRunnerVersion = "await-runner-healthy-failfast-unhealthy-v1"

// why: Existing histories must retain the aggregate-status fail-fast behavior they recorded.
const processReadinessPolicyVersion = "await-runner-healthy-process-readiness-v1"

type Mode string

const (
	ModeStartup       Mode = "startup"
	ModeRequireActive Mode = "require-active"
)

type Signal struct {
	InstallID      string `json:"install_id"`
	WorkflowStepID string `json:"workflow_step_id"`
	Mode           Mode   `json:"mode"`

	v *validator.Validate
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithParams         = (*Signal)(nil)
	_ signal.SignalWithStepContext    = (*Signal)(nil)
	_ signal.SignalWithAutoRetry      = (*Signal)(nil)
	_ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
	_ signal.SignalWithSkippable      = (*Signal)(nil)
)

func (s *Signal) AutoRetry() bool { return true }
func (s *Signal) Skippable() bool { return false }

func (s *Signal) MaxAutoRetries(ctx workflow.Context) int { return 0 }

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
	if s.Mode != "" && s.Mode != ModeStartup && s.Mode != ModeRequireActive {
		return errors.Errorf("unsupported runner readiness mode %q", s.Mode)
	}

	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	skipDisabled := workflow.GetVersion(ctx, skipDisabledRunnerVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	if skipDisabled && runnerDisabled(runner) {
		if s.WorkflowStepID != "" {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: s.WorkflowStepID,
				Status: app.CompositeStatus{
					Status:                 app.StatusAutoSkipped,
					StatusHumanDescription: "runner is disabled by the install stack",
				},
			}); err != nil {
				return errors.Wrap(err, "unable to mark step auto-skipped for disabled runner")
			}
		}
		return nil
	}

	failFast := workflow.GetVersion(ctx, failFastUnhealthyRunnerVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	processReadiness := workflow.GetVersion(ctx, processReadinessPolicyVersion, workflow.DefaultVersion, 1)

	processType := app.InstallProcessForRunnerGroupType(runner.RunnerGroup.Type)
	if processType == app.RunnerProcessTypeUnknown {
		return errors.Errorf("unsupported runner group type %s for health checking", runner.RunnerGroup.Type)
	}

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   runner.ID,
			StepTargetType: "runners",
		}); err != nil {
			return errors.Wrap(err, "unable to update workflow step target")
		}
	}

	processReq := activities.GetCurrentRunnerProcessRequest{
		RunnerID:    runner.ID,
		ProcessType: processType,
	}

	if processReadiness == workflow.DefaultVersion {
		if failFast && runnerCannotBecomeHealthy(runner.Status) {
			return errors.Errorf("runner is %s and cannot process jobs; check the runner status and try again", runner.Status)
		}
		return s.awaitActiveProcess(ctx, processReq)
	}

	if s.Mode == "" {
		if failFast && runnerCannotBecomeHealthy(runner.Status) {
			return errors.Errorf("runner is %s and cannot process jobs; check the runner status and try again", runner.Status)
		}
		return s.awaitActiveProcess(ctx, processReq)
	}
	if s.Mode == ModeRequireActive {
		return requireActiveProcess(ctx, processReq)
	}
	return s.awaitActiveProcess(ctx, processReq)
}

func requireActiveProcess(ctx workflow.Context, processReq activities.GetCurrentRunnerProcessRequest) error {
	process, err := activities.AwaitGetCurrentRunnerProcess(ctx, processReq)
	if err != nil {
		var appErr *temporal.ApplicationError
		if errors.As(err, &appErr) && appErr.Type() == "not found" {
			return errors.Wrap(poll.NonRetryableError, "runner has no active process; check the runner status and try again")
		}
		return errors.Wrap(err, "unable to get current runner process")
	}
	if process.ProcessStatus() != app.RunnerProcessStatusActive {
		return errors.Wrapf(poll.NonRetryableError, "runner process is %s; check the runner status and try again", process.ProcessStatus())
	}
	return nil
}

func (s *Signal) awaitActiveProcess(ctx workflow.Context, processReq activities.GetCurrentRunnerProcessRequest) error {
	if err := poll.Poll(ctx, s.v, poll.PollOpts{
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

			l.Debug("checking runner process status again", zap.Duration("next_check_in", dur))
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "runner process did not become healthy")
	}

	return nil
}

func runnerDisabled(runner *app.Runner) bool {
	return runner.Status == app.RunnerStatusDisabled || runner.StatusV2.Status == app.Status(app.RunnerStatusDisabled)
}

func runnerCannotBecomeHealthy(status app.RunnerStatus) bool {
	return status == app.RunnerStatusOffline || status == app.RunnerStatusError
}
