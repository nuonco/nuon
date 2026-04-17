package sandboxmode

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Signal wraps a real signal and checks for a SandboxModeSignalConfig.
// Config is fetched on the first call to Validate or Execute and cached.
// If a config exists and is enabled, sandbox behavior is applied instead
// of the real signal logic.
type Signal struct {
	signal.Signal
}

// WrapSignal wraps the given signal with sandbox-mode checking.
func WrapSignal(inner signal.Signal) *Signal {
	return &Signal{Signal: inner}
}
