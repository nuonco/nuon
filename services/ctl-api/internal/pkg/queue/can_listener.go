package queue

import (
	"math/rand"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const (
	canMinInterval = 2 * time.Minute
	canMaxJitter   = 3 * time.Minute // interval is 2-5 minutes
	canStartJitter = 60              // seconds of initial jitter
	canHistoryMax  = 10000
)

func (q *queue) startCANListener(ctx workflow.Context) {
	workflow.Go(ctx, func(gCtx workflow.Context) {
		l, _ := log.WorkflowLogger(gCtx)

		// Stagger startup across queues to avoid thundering herd.
		jitter := time.Duration(rand.Intn(canStartJitter)) * time.Second
		if err := workflow.Sleep(gCtx, jitter); err != nil {
			return
		}

		for {
			interval := canMinInterval + time.Duration(rand.Intn(int(canMaxJitter.Seconds())))*time.Second
			if err := workflow.Sleep(gCtx, interval); err != nil {
				return
			}

			// Check 1: Temporal suggests CAN due to large history.
			if workflow.GetInfo(gCtx).GetCurrentHistoryLength() > canHistoryMax {
				if l != nil {
					l.Info("history length exceeded threshold, triggering continue-as-new",
						zap.Int("history_length", workflow.GetInfo(gCtx).GetCurrentHistoryLength()))
				}
				q.restarted = true
				return
			}

			// Check 2: continue_as_new_requested set in queue metadata.
			requested, err := activities.AwaitCheckCANRequested(gCtx, activities.CheckCANRequestedRequest{
				QueueID: q.queueID,
			})
			if err != nil {
				if l != nil {
					l.Warn("unable to check CAN requested", zap.Error(err))
				}
				continue
			}

			if requested {
				if l != nil {
					l.Info("continue-as-new requested via metadata")
				}
				q.restarted = true
				return
			}
		}
	})
}
