package executeactionworkflowcron

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
)

const SignalType signal.SignalType = "execute_action_workflow_cron"

type Signal struct {
	InstallID               string `json:"install_id"`
	InstallActionWorkflowID string `json:"install_action_workflow_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.InstallActionWorkflowID == "" {
		return errors.New("install_action_workflow_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return signalsactivities.AwaitPkgSignalsSendInstallsSignal(ctx, &signalsactivities.SendSignalRequest[*installsignals.Signal]{
		ID: s.InstallID,
		Signal: &installsignals.Signal{
			Type: installsignals.OperationExecuteActionWorkflow,
			InstallActionWorkflowTrigger: installsignals.InstallActionWorkflowTriggerSubSignal{
				InstallActionWorkflowID: s.InstallActionWorkflowID,
				TriggerType:             app.ActionWorkflowTriggerTypeCron,
				TriggeredByType:         "cron",
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE": "cron",
				},
			},
		},
	})
}
