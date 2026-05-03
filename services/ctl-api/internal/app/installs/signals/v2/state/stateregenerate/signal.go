package stateregenerate

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statesignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

const SignalType signal.SignalType = "state-regenerate"

type Signal struct {
	InstallID       string
	Targets         []state.PartialTarget
	ForceAll        bool
	TriggeredByID   string
	TriggeredByType app.InstallStateGenerateSource
}

var (
	_ signal.Signal              = &Signal{}
	_ signal.SignalWithAutoRetry = (*Signal)(nil)
)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if len(s.Targets) == 0 {
		return fmt.Errorf("targets is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	_, err := statesignals.Regenerate(ctx, &state.ExecuteRegenerationRequest{
		InstallID:       s.InstallID,
		Targets:         s.Targets,
		ForceAll:        s.ForceAll,
		TriggeredByID:   s.TriggeredByID,
		TriggeredByType: s.TriggeredByType,
	})
	return err
}
