package rolechange

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "role-change"

const installSignalsQueueName = "install-signals"

type Signal struct {
	InstallID      string `json:"install_id"`
	RoleName       string `json:"role_name"`
	RoleType       string `json:"role_type"`
	ChangeType     string `json:"change_type"`
	RoleID         string `json:"role_id"`
	InstallRolesID string `json:"install_roles_id"`
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
		Operation: "role-change",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.RoleName == "" {
		return errors.New("role_name is required")
	}
	if s.ChangeType == "" {
		return errors.New("change_type is required")
	}
	return nil
}

func (s *Signal) Execute(_ workflow.Context) error {
	return nil
}
