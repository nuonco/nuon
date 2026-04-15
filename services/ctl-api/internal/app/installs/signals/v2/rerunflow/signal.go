package rerunflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType qsignal.SignalType = "rerun-flow"

type Signal struct {
	signal.Hooks
	InstallID          string                     `json:"install_id"`
	InstallWorkflowID  string                     `json:"install_workflow_id"`
	RerunConfiguration signals.RerunConfiguration `json:"rerun_configuration"`
}

var _ qsignal.Signal = (*Signal)(nil)

func (s *Signal) Type() qsignal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	if s.InstallWorkflowID == "" {
		return errors.New("install_workflow_id is required")
	}

	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return s.rerunFlow(ctx)
}
