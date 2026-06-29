package inputsupdated

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "inputs-updated"

const installSignalsQueueName = "install-signals"

type Signal struct {
	InstallID   string   `json:"install_id"`
	ChangedKeys []string `json:"changed_keys,omitempty"`
	AddedKeys   []string `json:"added_keys,omitempty"`
	RemovedKeys []string `json:"removed_keys,omitempty"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }
func (s *Signal) AutoRetry() bool         { return true }
func (s *Signal) MaxRetries() int         { return 5 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	return signal.SignalLifecycleContext{
		InstallID: installID,
		Operation: "inputs-updated",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	return nil
}

func (s *Signal) Execute(_ workflow.Context) error {
	return nil
}
