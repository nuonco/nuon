package sandboxmode

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

// fetchConfig loads the sandbox config once and caches it.
func (s *Signal) fetchConfig(ctx workflow.Context) *app.SandboxModeSignalConfig {
	cfg, err := activities.AwaitGetSandboxSignalConfigBySignalType(ctx, string(s.Signal.Type()))
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return nil
		}
	}

	if !cfg.Enabled {
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
