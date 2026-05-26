package queue

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

func (q *queue) startHintListener(ctx workflow.Context) {
	workflow.Go(ctx, func(gCtx workflow.Context) {
		l, _ := log.WorkflowLogger(gCtx)
		for {
			if err := workflow.Sleep(gCtx, 30*time.Second); err != nil {
				return
			}

			hint, err := activities.AwaitCheckRestartHint(gCtx, activities.CheckRestartHintRequest{
				QueueID: q.queueID,
			})
			if err != nil {
				if generics.IsGormErrRecordNotFound(err) {
					if l != nil {
						l.Warn("queue not found during restart hint check, stopping workflow", zap.String("queue-id", q.queueID))
					}
					q.stopped = true
					return
				}
				if l != nil {
					l.Warn("unable to check restart hint", zap.Error(err))
				}
				continue
			}

			if hint {
				if l != nil {
					l.Info("restart hint detected, triggering continue-as-new")
				}
				q.setStatus(gCtx, l, QueueStatusRestartAccepted)
				q.restarted = true
				return
			}
		}
	})
}
