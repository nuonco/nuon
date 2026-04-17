package sandboxmode

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Signal wraps a real signal and checks for a SandboxModeSignalConfig
// during Validate and Execute. If a config exists for the signal type,
// sandbox behavior is applied instead of the real signal logic.
type Signal struct {
	signal.Signal
}

// WrapSignal wraps the given signal with sandbox-mode checking.
func WrapSignal(inner signal.Signal) *Signal {
	return &Signal{Signal: inner}
}

func (s *Signal) fetchConfig(ctx workflow.Context) *app.SandboxModeSignalConfig {
	cfg, err := activities.AwaitGetSandboxSignalConfigBySignalType(ctx, string(s.Signal.Type()))
	if err != nil || cfg == nil || !cfg.Enabled {
		return nil
	}
	return cfg
}

func (s *Signal) applyConfig(ctx workflow.Context, cfg *app.SandboxModeSignalConfig) error {
	if cfg.Panic {
		panic("sandbox signal config: panic requested for " + string(s.Signal.Type()))
	}
	if cfg.DeadlockSleep > 0 {
		time.Sleep(cfg.DeadlockSleep)
		return errors.New("sandbox signal config: deadlock sleep expired")
	}
	if cfg.WorkflowSleep > 0 {
		_ = workflow.Sleep(ctx, cfg.WorkflowSleep)
		return nil
	}
	if cfg.Error != "" {
		return errors.New(cfg.Error)
	}
	return nil
}
