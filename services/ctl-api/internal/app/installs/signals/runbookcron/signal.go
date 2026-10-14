package runbookcron

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "runbook-cron"

type Signal struct {
	InstallID        string `json:"install_id"`
	InstallRunbookID string `json:"install_runbook_id"`
	RunbookConfigID  string `json:"runbook_config_id"`
	TriggeredByID    string `json:"triggered_by_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "runbook-cron",
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.InstallRunbookID == "" {
		return fmt.Errorf("install_runbook_id is required")
	}
	if s.RunbookConfigID == "" {
		return fmt.Errorf("runbook_config_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if disabled, err := activities.AwaitIsRunnerDisabled(ctx, activities.IsRunnerDisabledRequest{
		InstallID: s.InstallID,
	}); err != nil {
		return fmt.Errorf("unable to check whether the install runner is disabled: %w", err)
	} else if disabled.Disabled {
		workflow.GetLogger(ctx).Info("install runner is disabled, skipping runbook cron",
			"install_id", s.InstallID)
		return nil
	}

	if err := activities.AwaitTriggerRunbookCron(ctx, activities.TriggerRunbookCronRequest{
		InstallRunbookID: s.InstallRunbookID,
		RunbookConfigID:  s.RunbookConfigID,
		TriggeredByID:    s.TriggeredByID,
		IdempotencyKey:   fmt.Sprintf("runbook-cron:%s", workflow.GetInfo(ctx).WorkflowExecution.ID),
	}); err != nil {
		return fmt.Errorf("unable to trigger runbook cron run: %w", err)
	}
	return nil
}
