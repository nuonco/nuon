package appconfigsynced

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-config-synced"

type Signal struct {
	AppID       string `json:"app_id" validate:"required"`
	AppBranchID string `json:"app_branch_id,omitempty"`
	AppName     string `json:"app_name,omitempty"`
	BranchName  string `json:"branch_name,omitempty"`
	ActorEmail  string `json:"actor_email,omitempty"`
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
	ctx := signal.SignalLifecycleContext{
		Operation: "app-config-synced",
		OwnerID:   s.AppID,
		OwnerType: "apps",
		Metadata: map[string]any{
			"app_name":    s.AppName,
			"branch_name": s.BranchName,
			"actor_email": s.ActorEmail,
		},
	}
	if s.AppBranchID != "" {
		ctx.OwnerID = s.AppBranchID
		ctx.OwnerType = "app_branches"
	}
	return ctx
}

func (s *Signal) Validate(_ workflow.Context) error {
	return nil
}

func (s *Signal) Execute(_ workflow.Context) error {
	return nil
}
