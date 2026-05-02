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

	var actionEmitters, driftEmitters []app.QueueEmitter
	for _, em := range existingEmitters {
		switch {
		case strings.HasPrefix(em.Name, actionCronEmitterPrefix):
			actionEmitters = append(actionEmitters, em)
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
