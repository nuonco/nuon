package handler

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	// Periodically check that the queue signal still exists and hasn't expired.
	// If deleted or expired, set h.stopped so the Await below unblocks and the
	// handler workflow terminates cleanly.
	workflow.Go(ctx, func(gCtx workflow.Context) {
		for {
			if err := workflow.Sleep(gCtx, 1*time.Minute); err != nil {
				return
			}
			if h.finished || h.stopped || h.restarted {
				return
			}
			qs, err := activities.AwaitGetQueueSignalByQueueSignalID(gCtx, h.queueSignalID)
			if err != nil {
				if dbgenerics.IsGormErrRecordNotFound(err) {
					l.Warn("queue signal deleted, terminating orphaned handler")
					if h.mw != nil {
						h.mw.Incr("queue.handler.signal_not_found", metrics.ToTags(map[string]string{
							"queue_id": h.queueID,
						}))
					}
					h.stopped = true
					return
				}
				continue
			}

			// If the signal has an expiry and it has passed, mark it as
			// expired in the DB so callers see a terminal status.
			if qs.ExpiresAt != nil && workflow.Now(gCtx).After(*qs.ExpiresAt) {
				l.Warn("queue signal expired, terminating handler")
				if h.mw != nil {
					h.mw.Incr("queue.handler.signal_expired", metrics.ToTags(map[string]string{
						"queue_id":    h.queueID,
						"signal_type": string(qs.Type),
					}))
				}
				_ = statusactivities.AwaitUpdateQueueSignalStatusV2(gCtx, statusactivities.UpdateQueueSignalStatusV2Request{
					QueueSignalID:     h.queueSignalID,
					Status:            app.StatusError,
					StatusDescription: "signal expired without being executed",
				})
				h.stopped = true
				return
			}
		}
	})

	// execute the handler and handle a restart or stop
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(h.stopped, h.restarted, h.finished)
	}); err != nil {
		return false, err
	}
	if h.restarted {
		return false, nil
	}
	if h.stopped {
		return true, nil
	}

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
