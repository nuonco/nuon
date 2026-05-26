package appconfigupdated

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "appconfig-updated"

const (
	actionCronEmitterPrefix = "action-cron-"
	driftEmitterPrefix      = "drift-"
	// driftSandboxEmitterPrefix is the per-install sandbox drift cron prefix.
	// IMPORTANT: this string starts with the more generic `drift-` prefix —
	// the splitting loop in Execute() must check for `drift-sandbox-` BEFORE
	// the bare `drift-` case or sandbox emitters get classified as
	// per-component drift emitters and reconciled against the wrong list.
	driftSandboxEmitterPrefix = "drift-sandbox-"
)

type Signal struct {
	InstallID string `json:"install_id"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "appconfig-updated",
	}
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}

	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("install not found: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	signalsQueue, err := activities.AwaitGetInstallSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install signals queue: %w", err)
	}

	actionCronQueue, err := activities.AwaitGetInstallActionCronSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install action cron signals queue: %w", err)
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// Fetch existing emitters from the signals queue (drift emitters live here)
	signalsEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, signalsQueue.ID)
	if err != nil {
		return fmt.Errorf("unable to get signals queue emitters: %w", err)
	}

	// Fetch existing emitters from the action cron queue
	actionCronEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, actionCronQueue.ID)
	if err != nil {
		return fmt.Errorf("unable to get action cron queue emitters: %w", err)
	}

	// Split signals queue emitters by prefix. Action cron emitters that were
	// previously on the signals queue are also collected for cleanup.
	var legacyActionEmitters, driftEmitters, driftSandboxEmitters []app.QueueEmitter
	for _, em := range signalsEmitters {
		switch {
		case strings.HasPrefix(em.Name, actionCronEmitterPrefix):
			// Legacy: action cron emitters that lived on the signals queue
			// before they got their own queue. Clean them up.
			legacyActionEmitters = append(legacyActionEmitters, em)
		// Check the more specific `drift-sandbox-` prefix BEFORE the bare
		// `drift-` case — otherwise sandbox emitters would be swept into
		// the per-component drift bucket.
		case strings.HasPrefix(em.Name, driftSandboxEmitterPrefix):
			driftSandboxEmitters = append(driftSandboxEmitters, em)
		case strings.HasPrefix(em.Name, driftEmitterPrefix):
			driftEmitters = append(driftEmitters, em)
		}
	}

	// Clean up legacy action cron emitters from the signals queue.
	stopAndDeleteEmitters(ctx, l, legacyActionEmitters)

	// Collect action cron emitters from their dedicated queue.
	var actionEmitters []app.QueueEmitter
	for _, em := range actionCronEmitters {
		if strings.HasPrefix(em.Name, actionCronEmitterPrefix) {
			actionEmitters = append(actionEmitters, em)
		}
	}

	if err := s.reconcileActionCronEmitters(ctx, l, install, appCfg, actionCronQueue, actionEmitters); err != nil {
		return fmt.Errorf("unable to reconcile action cron emitters: %w", err)
	}

	if err := s.reconcileDriftEmitters(ctx, l, install, appCfg, signalsQueue, driftEmitters); err != nil {
		return fmt.Errorf("unable to reconcile drift emitters: %w", err)
	}

	if err := s.reconcileDriftSandboxEmitter(ctx, l, install, signalsQueue, driftSandboxEmitters); err != nil {
		return fmt.Errorf("unable to reconcile sandbox drift emitter: %w", err)
	}

	return nil
}

// stopAndDeleteEmitters stops and deletes a list of emitters. Errors are logged but not fatal.
func stopAndDeleteEmitters(ctx workflow.Context, l interface{ Warn(string, ...interface{}) }, emitters []app.QueueEmitter) {
	for _, em := range emitters {
		if _, err := emitterclient.AwaitStopEmitter(ctx, em.ID); err != nil {
			l.Warn("unable to stop emitter",
				zap.String("emitter_id", em.ID),
				zap.Error(err))
		}
		if err := emitterclient.AwaitDeleteEmitter(ctx, em.ID); err != nil {
			l.Warn("unable to delete emitter",
				zap.String("emitter_id", em.ID),
				zap.Error(err))
		}
	}
}
