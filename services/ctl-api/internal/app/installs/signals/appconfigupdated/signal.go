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
	InstallID      string `json:"install_id"`
	OldAppConfigID string `json:"old_app_config_id,omitempty"`
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

	if s.OldAppConfigID != "" && s.OldAppConfigID != install.AppConfigID {
		if err := activities.AwaitMigrateInstallInputs(ctx, &activities.MigrateInstallInputsInput{
			InstallID:      s.InstallID,
			OldAppConfigID: s.OldAppConfigID,
			NewAppConfigID: install.AppConfigID,
		}); err != nil {
			l.Warn("unable to migrate install inputs",
				zap.String("install_id", s.InstallID),
				zap.Error(err))
		}
	}

	// Best-effort: default-label reconciliation must not fail config updates.
	if err := activities.AwaitApplyAppDefaultLabels(ctx, &activities.ApplyAppDefaultLabelsRequest{
		InstallID: s.InstallID,
	}); err != nil {
		l.Warn("unable to apply app default labels",
			zap.String("install_id", s.InstallID),
			zap.Error(err))
	}

	// A config change can move config-derived state (.nuon.app, migrated inputs)
	// without touching the default-label set, and ApplyAppDefaultLabels only
	// renders when that set changed. Best-effort re-render so templates pick up
	// the new config; reads regenerate the partials marked stale above.
	if err := activities.AwaitRenderInstallLabels(ctx, &activities.RenderInstallLabelsRequest{
		InstallID: s.InstallID,
	}); err != nil {
		l.Warn("unable to render install label templates",
			zap.String("install_id", s.InstallID),
			zap.Error(err))
	}

	signalsQueue, err := activities.AwaitGetInstallSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install signals queue: %w", err)
	}

	actionCronQueue, err := activities.AwaitGetInstallActionCronSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install action cron signals queue: %w", err)
	}

	driftCronQueue, err := activities.AwaitGetInstallDriftCronSignalsQueueByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install drift cron signals queue: %w", err)
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	signalsEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, signalsQueue.ID)
	if err != nil {
		return fmt.Errorf("unable to get signals queue emitters: %w", err)
	}

	actionCronEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, actionCronQueue.ID)
	if err != nil {
		return fmt.Errorf("unable to get action cron queue emitters: %w", err)
	}

	driftCronEmitters, err := emitterclient.AwaitGetEmittersByQueueID(ctx, driftCronQueue.ID)
	if err != nil {
		return fmt.Errorf("unable to get drift cron queue emitters: %w", err)
	}

	// Legacy: cron emitters that lived on the signals queue before they got
	// their own dedicated queues. Clean them up.
	var legacyEmitters []app.QueueEmitter
	for _, em := range signalsEmitters {
		if strings.HasPrefix(em.Name, actionCronEmitterPrefix) ||
			strings.HasPrefix(em.Name, driftSandboxEmitterPrefix) ||
			strings.HasPrefix(em.Name, driftEmitterPrefix) {
			legacyEmitters = append(legacyEmitters, em)
		}
	}
	stopAndDeleteEmitters(ctx, l, legacyEmitters)

	var actionEmitters []app.QueueEmitter
	for _, em := range actionCronEmitters {
		if strings.HasPrefix(em.Name, actionCronEmitterPrefix) {
			actionEmitters = append(actionEmitters, em)
		}
	}

	// Check the more specific `drift-sandbox-` prefix BEFORE the bare `drift-`
	// case — otherwise sandbox emitters get swept into the per-component bucket.
	var driftEmitters, driftSandboxEmitters []app.QueueEmitter
	for _, em := range driftCronEmitters {
		switch {
		case strings.HasPrefix(em.Name, driftSandboxEmitterPrefix):
			driftSandboxEmitters = append(driftSandboxEmitters, em)
		case strings.HasPrefix(em.Name, driftEmitterPrefix):
			driftEmitters = append(driftEmitters, em)
		}
	}

	if err := s.reconcileActionCronEmitters(ctx, l, install, appCfg, actionCronQueue, actionEmitters); err != nil {
		return fmt.Errorf("unable to reconcile action cron emitters: %w", err)
	}

	if err := s.reconcileDriftEmitters(ctx, l, install, appCfg, driftCronQueue, driftEmitters); err != nil {
		return fmt.Errorf("unable to reconcile drift emitters: %w", err)
	}

	if err := s.reconcileDriftSandboxEmitter(ctx, l, install, driftCronQueue, driftSandboxEmitters); err != nil {
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
