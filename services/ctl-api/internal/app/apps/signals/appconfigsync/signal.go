package appconfigsync

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-config-sync"

// Signal applies the intermediate config stored on an app config. Backs
// POST /v1/apps/:app_id/configs/:config_id/sync (the CLI sync path).
type Signal struct {
	AppID       string `json:"app_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`

	// AccountID advances the requesting account's onboarding journey.
	AccountID string `json:"account_id,omitempty"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		Operation: "app-config-sync",
		OwnerID:   s.AppID,
		OwnerType: "apps",
		Metadata: map[string]any{
			"app_config_id": s.AppConfigID,
		},
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	if s.AppConfigID == "" {
		return fmt.Errorf("app_config_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	result, err := AwaitApplyAppConfig(ctx, ApplyAppConfigRequest{
		Req: &ApplyAppConfigInput{
			AppID:       s.AppID,
			AppConfigID: s.AppConfigID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to sync app config: %w", err)
	}

	if len(result.ComponentIDsToBuild) > 0 {
		if err := AwaitDispatchComponentBuilds(ctx, DispatchComponentBuildsRequest{
			Req: &DispatchComponentBuildsInput{
				Components: result.ComponentIDsToBuild,
			},
		}); err != nil {
			return fmt.Errorf("unable to dispatch component builds: %w", err)
		}
	}

	// install rollout + journey step, previously side effects of the CLI's
	// final PATCH-to-active
	if err := AwaitFinalizeAppConfigSync(ctx, FinalizeAppConfigSyncRequest{
		Req: &FinalizeAppConfigSyncInput{
			AppID:       s.AppID,
			AppConfigID: s.AppConfigID,
			AccountID:   s.AccountID,
		},
	}); err != nil {
		return fmt.Errorf("unable to finalize app config sync: %w", err)
	}

	return nil
}
