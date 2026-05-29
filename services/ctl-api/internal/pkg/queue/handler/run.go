package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/workflowmanager"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (h *handler) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	// Check that the signal still exists before doing any work.
	// If it was deleted, terminate the workflow without continue-as-new.
	if _, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID); err != nil {
		if dbgenerics.IsGormErrRecordNotFound(err) {
			l.Warn("queue signal not found, terminating handler")
			return true, nil
		}
		return false, errors.Wrap(err, "unable to fetch queue signal")
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

	// Start the lifecycle manager to periodically check that the queue signal
	// still exists and hasn't expired. Sets mgr.Stopped when the entity is
	// gone or expired, which unblocks the Await below.
	mgr := workflowmanager.New(
		workflowmanager.WithCheckInterval(1*time.Minute),
		workflowmanager.WithAliveChecker(func(gCtx workflow.Context) (bool, error) {
			_, err := activities.AwaitGetQueueSignalByQueueSignalID(gCtx, h.queueSignalID)
			if err != nil {
				if dbgenerics.IsGormErrRecordNotFound(err) {
					l.Warn("queue signal deleted, terminating orphaned handler")
					return false, nil
				}
				return true, nil // transient error, keep going
			}
			return true, nil
		}),
		workflowmanager.WithExpiryChecker(func(gCtx workflow.Context) (*time.Time, error) {
			qs, err := activities.AwaitGetQueueSignalByQueueSignalID(gCtx, h.queueSignalID)
			if err != nil {
				return nil, err
			}
			return qs.ExpiresAt, nil
		}),
		workflowmanager.WithOnStopped(func(gCtx workflow.Context) {
			_ = statusactivities.AwaitUpdateQueueSignalStatusV2(gCtx, statusactivities.UpdateQueueSignalStatusV2Request{
				QueueSignalID:     h.queueSignalID,
				Status:            app.StatusError,
				StatusDescription: "signal expired or deleted",
			})
		}),
	)
	mgr.Start(ctx)

	// execute the handler and handle a restart or stop
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(mgr.Stopped, mgr.Restarted, h.finished)
	}); err != nil {
		return false, err
	}
	if mgr.Restarted {
		return false, nil
	}
	if mgr.Stopped {
		// Entity was deleted or expired. Send callbacks so waiting callers unblock.
		h.sendCompletionCallbacks(ctx)
		return true, nil
	}

	// Signal completed — send completion callbacks to unblock queue and parent.
	// Always call sendCompletionCallbacks (it reloads from DB) so that callbacks
	// added dynamically by EnsureSignal after init are picked up.
	h.sendCompletionCallbacks(ctx)

	// Once execution has completed, keep the workflow alive for a cache period
	// so that subsequent signals can reuse it via update-with-start.
	cacheDur := signal.DefaultSleepAfter
	if sa, ok := h.sig.(signal.SleepAfter); ok {
		cacheDur = sa.SleepAfter()
	}
	if cacheDur > 0 {
		l.Debug("handler finished, caching workflow")
		_ = workflow.Sleep(ctx, cacheDur)
	}

	return true, nil
}
