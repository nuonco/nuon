package updated

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-updated"

type Signal struct {
	AppBranchID       string `json:"app_branch_id"`
	AppBranchConfigID string `json:"app_branch_config_id"`
}

var (
	_ signal.Signal               = (*Signal)(nil)
	_ signal.SignalWithAutoRetry  = (*Signal)(nil)
	_ signal.SignalWithMaxRetries = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }
func (s *Signal) AutoRetry() bool         { return true }
func (s *Signal) MaxRetries() int         { return 5 }

func (s *Signal) Validate(_ workflow.Context) error {
	if s.AppBranchID == "" {
		return fmt.Errorf("app_branch_id is required")
	}
	if s.AppBranchConfigID == "" {
		return fmt.Errorf("app_branch_config_id is required")
	}
	return nil
}
