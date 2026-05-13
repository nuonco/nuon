package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/can"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

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

	if err := signal.RegisterUpdateHandlers(h.sig, ctx); err != nil {
		return false, errors.Wrap(err, "unable to register signal update handlers")
	}

	l.Debug("handler is ready")
	h.ready = true

	// execute the handler and handle a restart or stop
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(h.stopped, h.restarted, h.finished) || can.ShouldContinueAsNew(ctx)
	}); err != nil {
		return false, err
	}
	if h.restarted {
		return false, nil
	}
	if h.stopped {
		return true, nil
	}
	// History-driven CAN before terminal flags fire — signal state lives in the DB
	// and re-initializes on the next run, so this is safe mid-execution.
	if !h.finished && can.ShouldContinueAsNew(ctx) {
		l.Info("history-driven continue-as-new",
			zap.Int("history_length", workflow.GetInfo(ctx).GetCurrentHistoryLength()),
			zap.Bool("server_suggested", workflow.GetInfo(ctx).GetContinueAsNewSuggested()),
		)
		return false, nil
	}

	// Once execution has completed, keep the workflow alive for a cache period
	// so that subsequent signals can reuse it via update-with-start. Skip the
	// cache sleep if history is already large — better to terminate now than
	// pad more events onto a fat workflow.
	cacheDur := signal.DefaultSleepAfter
	if sa, ok := h.sig.(signal.SleepAfter); ok {
		cacheDur = sa.SleepAfter()
	}
	if cacheDur > 0 && !can.ShouldContinueAsNew(ctx) {
		l.Debug("handler finished, caching workflow")
		_ = workflow.Sleep(ctx, cacheDur)
	}

	return true, nil
}
