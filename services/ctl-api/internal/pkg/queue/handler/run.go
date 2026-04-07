package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const defaultHandlerGracePeriod = 15 * time.Minute

func (h *handler) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	if err := h.registerHandlers(ctx); err != nil {
		return false, err
	}

	if err := h.initializeState(ctx); err != nil {
		return false, errors.Wrap(err, "unable to initialize state")
	}

	l.Debug("handler is ready")
	h.ready = true

	// Phase 1: Wait for the handler to be stopped, restarted, finished, or put to sleep.
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(h.stopped, h.restarted, h.finished, h.sleeping)
	}); err != nil {
		return false, err
	}

	if h.restarted {
		return false, nil
	}
	if h.stopped {
		return true, nil
	}
	if h.sleeping {
		return h.dormantSleep(ctx, l)
	}

	// Phase 2: Handler finished execution — enter grace period.
	gracePeriod := h.cfg.QueueHandlerGracePeriod
	if gracePeriod == 0 {
		gracePeriod = defaultHandlerGracePeriod
	}
	workflow.Go(ctx, func(gCtx workflow.Context) {
		_ = workflow.Sleep(gCtx, gracePeriod)
		h.graceExpired = true
	})

	l.Debug("handler finished, entering grace period")
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(h.stopped, h.restarted, h.sleeping, h.graceExpired)
	}); err != nil {
		return false, err
	}

	if h.sleeping {
		return h.dormantSleep(ctx, l)
	}
	if h.restarted {
		return false, nil
	}

	// graceExpired or stopped — terminate
	l.Debug("handler grace period expired, terminating")
	return true, nil
}

// dormantSleep puts the handler into a dormant state where it waits
// indefinitely until woken via the wake update handler. When woken,
// it returns false to trigger a continue-as-new.
func (h *handler) dormantSleep(ctx workflow.Context, l *zap.Logger) (bool, error) {
	l.Debug("handler entering dormant sleep")
	if err := workflow.Await(ctx, func() bool {
		return h.woken
	}); err != nil {
		return false, err
	}
	l.Debug("handler woken from dormant sleep")
	return false, nil
}
