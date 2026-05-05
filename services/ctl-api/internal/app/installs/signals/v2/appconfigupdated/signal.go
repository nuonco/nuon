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

	queue, err := activities.AwaitGetInstallSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install signals queue: %w", err)
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// Fetch existing emitters and split by prefix
	existingEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, queue.ID)
	if err != nil {
		return fmt.Errorf("unable to get existing emitters: %w", err)
	}

	var actionEmitters, driftEmitters, driftSandboxEmitters []app.QueueEmitter
	for _, em := range existingEmitters {
		switch {
		case strings.HasPrefix(em.Name, actionCronEmitterPrefix):
			actionEmitters = append(actionEmitters, em)
		// Check the more specific `drift-sandbox-` prefix BEFORE the bare
		// `drift-` case — otherwise sandbox emitters would be swept into
		// the per-component drift bucket.
		case strings.HasPrefix(em.Name, driftSandboxEmitterPrefix):
			driftSandboxEmitters = append(driftSandboxEmitters, em)
		case strings.HasPrefix(em.Name, driftEmitterPrefix):
			driftEmitters = append(driftEmitters, em)
		}
	}

	if err := s.reconcileActionCronEmitters(ctx, l, install, appCfg, queue, actionEmitters); err != nil {
		return fmt.Errorf("unable to reconcile action cron emitters: %w", err)
	}

	if err := s.reconcileDriftEmitters(ctx, l, install, appCfg, queue, driftEmitters); err != nil {
		return fmt.Errorf("unable to reconcile drift emitters: %w", err)
	}

	if err := s.reconcileDriftSandboxEmitter(ctx, l, install, queue, driftSandboxEmitters); err != nil {
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
