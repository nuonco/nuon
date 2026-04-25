package state

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const stateManagerIdleTimeout = 30 * time.Minute

type stateManager struct {
	installID string
	state     *StateManagerState
	workflows *Workflows

	ready     bool
	stopped   bool
	restarted bool

	// pendingPartials accumulates partials that need regeneration from hints.
	pendingPartials map[PartialName]bool

	// forceRegenerate flags a full rebuild on next cycle.
	forceRegenerate bool
}

func (sm *stateManager) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	l.Info("state-manager: verifying install exists", zap.String("install_id", sm.installID))
	if _, err := activities.AwaitGetByInstallID(ctx, sm.installID); err != nil {
		return true, errors.Wrap(err, "install not found, terminating state-manager")
	}

	l.Info("state-manager: registering handlers")
	if err := sm.registerHandlers(ctx); err != nil {
		return false, errors.Wrap(err, "unable to register handlers")
	}

	// First run: generate full state if no cached state exists.
	if sm.state.CachedState == nil {
		l.Info("state-manager: no cached state, performing full generation")
		if _, err := sm.executeRegeneration(ctx, allPartialsSet()); err != nil {
			return false, errors.Wrap(err, "unable to perform initial state generation")
		}
	}

	sm.ready = true
	l.Info("state-manager: ready")

	// Wait for work or exit conditions.
	if _, err := workflow.AwaitWithTimeout(ctx, stateManagerIdleTimeout, func() bool {
		return sm.restarted || sm.stopped || len(sm.pendingPartials) > 0 || sm.forceRegenerate
	}); err != nil {
		return false, err
	}

	if sm.stopped {
		l.Info("state-manager: stopped")
		return true, nil
	}
	if sm.restarted {
		l.Info("state-manager: restarting via continue-as-new")
		return false, nil
	}

	// Process pending work.
	if sm.forceRegenerate || len(sm.pendingPartials) > 0 {
		partials := sm.drainPendingPartials()
		if sm.forceRegenerate {
			partials = allPartialsSet()
			sm.forceRegenerate = false
		}
		if _, err := sm.executeRegeneration(ctx, partials); err != nil {
			l.Error("state-manager: regeneration failed", zap.Error(err))
			return false, errors.Wrap(err, "regeneration failed")
		}
	}

	// Continue-as-new after processing work (or idle timeout).
	return false, nil
}

func (sm *stateManager) drainPendingPartials() map[PartialName]bool {
	drained := sm.pendingPartials
	sm.pendingPartials = make(map[PartialName]bool)
	return drained
}
