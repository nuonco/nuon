package installstackrunstarted

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SignalType is the queue signal type for an install stack run started
// notification. Fired from the SDK-facing CreateInstallStackVersionRun
// endpoint after the run row is created. Notification-only: Execute is a
// no-op and exists solely so the queue dispatcher emits lifecycle events
// that webhook/Slack hooks can project to subscribers.
const SignalType signal.SignalType = "install-stack-run-started"

type Signal struct {
	InstallID              string `json:"install_id"`
	InstallStackID         string `json:"install_stack_id"`
	InstallStackVersionID  string `json:"install_stack_version_id"`
	InstallStackVersionRun string `json:"install_stack_version_run_id"`
	Kind                   string `json:"kind"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) AutoRetry() bool { return true }
func (s *Signal) MaxRetries() int { return 5 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	return signal.SignalLifecycleContext{
		InstallID: installID,
		// Operation carries the run kind (provision/reprovision/deprovision)
		// so the interests classifier can map this event onto the matching
		// install sub-op slug. The transition (started/succeeded/failed) is
		// derived from the SignalType, not Operation.
		Operation: s.Kind,
		OwnerID:   s.InstallID,
		OwnerType: "installs",
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.InstallStackVersionRun == "" {
		return errors.New("install_stack_version_run_id is required")
	}
	return nil
}

func (s *Signal) Execute(_ workflow.Context) error { return nil }
