package sandboxmode

import (
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SandboxSignal wraps a real signal and intercepts Validate/Execute
// based on the SandboxSignalConfig.
type SandboxSignal struct {
	inner  signal.Signal
	config *app.SandboxSignalConfig
}

var _ signal.Signal = (*SandboxSignal)(nil)

// WrapSignal creates a SandboxSignal wrapper around the real signal.
func WrapSignal(inner signal.Signal, config *app.SandboxSignalConfig) signal.Signal {
	return &SandboxSignal{
		inner:  inner,
		config: config,
	}
}

func (s *SandboxSignal) Type() signal.SignalType {
	return s.inner.Type()
}

func (s *SandboxSignal) Validate(ctx workflow.Context) error {
	return s.applyConfig(ctx)
}

func (s *SandboxSignal) Execute(ctx workflow.Context) error {
	return s.applyConfig(ctx)
}

func (s *SandboxSignal) applyConfig(ctx workflow.Context) error {
	if s.config.Panic {
		panic("sandbox signal config: panic requested for " + string(s.inner.Type()))
	}
	if s.config.Error != "" {
		return errors.New(s.config.Error)
	}
	if s.config.DeadlockSleep > 0 {
		// Real sleep — blocks the goroutine/activity, simulating a deadlock.
		// The Temporal activity timeout (30s) will eventually kill it.
		time.Sleep(s.config.DeadlockSleep)
		return errors.New("sandbox signal config: deadlock sleep expired")
	}
	if s.config.WorkflowSleep > 0 {
		_ = workflow.Sleep(ctx, s.config.WorkflowSleep)
		return nil
	}
	return nil
}

// Delegate optional interfaces so the handler framework still works.

func (s *SandboxSignal) AutoRetry() bool {
	if ar, ok := s.inner.(signal.SignalWithAutoRetry); ok {
		return ar.AutoRetry()
	}
	return false
}

func (s *SandboxSignal) LifecycleContext() signal.SignalLifecycleContext {
	if lc, ok := s.inner.(signal.SignalWithLifecycleContext); ok {
		return lc.LifecycleContext()
	}
	return signal.SignalLifecycleContext{}
}

func (s *SandboxSignal) SleepAfter() time.Duration {
	if sa, ok := s.inner.(signal.SleepAfter); ok {
		return sa.SleepAfter()
	}
	return signal.DefaultSleepAfter
}
