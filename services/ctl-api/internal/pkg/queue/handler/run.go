package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
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
	// We pass the fetched signal into initializeState to avoid a redundant DB fetch.
	qs, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(ctx, h.queueSignalID)
	if err != nil {
		if dbgenerics.IsGormErrRecordNotFound(err) {
			l.Warn("queue signal not found, terminating handler")
			return true, nil
		}
		return false, errors.Wrap(err, "unable to fetch queue signal")
	}

	if err := h.registerHandlers(ctx); err != nil {
		return false, err
	}

	if err := h.initializeState(ctx, qs); err != nil {
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
	//
	// The alive checker combines both existence and expiry checks in a single
	// DB query to avoid redundant round-trips.
	var mgrOpts []workflowmanager.Option
	// Handler signals have explicit callbacks for completion, so the alive
	// checker only needs to detect deletion/expiry. A longer interval reduces
	// local activity overhead for long-running signals.
	mgrOpts = append(mgrOpts, workflowmanager.WithCheckInterval(5*time.Minute))

	// don't continue-as-new mid-phase: it orphans the in-flight update and the
	// successor run fails the signal while the work is still alive.
	mgrOpts = append(mgrOpts, workflowmanager.WithDeferRestart(func() bool {
		return h.validating || h.executing
	}))

	// Create a temporal metrics writer for workflow size reporting.
	if h.mw != nil && h.v != nil {
		if tmw, err := tmetrics.New(h.v, tmetrics.WithMetricsWriter(h.mw)); err == nil {
			mgrOpts = append(mgrOpts, workflowmanager.WithMetricsWriter(tmw))
		}
	}

	mgrOpts = append(mgrOpts, workflowmanager.WithAliveChecker(func(gCtx workflow.Context) (bool, error) {
		qs, err := activities.LocalAwaitGetQueueSignalByQueueSignalID(gCtx, h.queueSignalID)
		if err != nil {
			if dbgenerics.IsGormErrRecordNotFound(err) {
				l.Warn("queue signal deleted, terminating orphaned handler")
				return false, nil
			}
			return true, nil // transient error, keep going
		}
		// Check expiry in the same call to avoid a second DB query.
		if qs.ExpiresAt != nil && workflow.Now(gCtx).After(*qs.ExpiresAt) {
			l.Warn("queue signal expired, stopping handler")
			return false, nil
		}
		return true, nil
	}))
	mgrOpts = append(mgrOpts, workflowmanager.WithOnStopped(func(gCtx workflow.Context) {
		_ = statusactivities.LocalAwaitUpdateQueueSignalStatusV2(gCtx, statusactivities.UpdateQueueSignalStatusV2Request{
			QueueSignalID:     h.queueSignalID,
			Status:            app.StatusError,
			StatusDescription: "signal expired or deleted",
		})
	}))
	mgr := workflowmanager.New(mgrOpts...)
	mgr.Start(ctx)

	// execute the handler and handle a restart or stop
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(mgr.Stopped, mgr.Restarted, h.finished)
	}); err != nil {
		return false, err
	}

	// drain in-flight phase updates first: ending the run mid-handler drops its deferred completion callback and wedges the dispatcher. bounded so a stuck handler can't leak the workflow
	if _, err := workflow.AwaitWithTimeout(ctx, callback.QuickTimeout, func() bool {
		return workflow.AllHandlersFinished(ctx)
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
