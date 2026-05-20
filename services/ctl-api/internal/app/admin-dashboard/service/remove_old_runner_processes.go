package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	removeoldrunnerprocesses "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/remove_old_runner_processes"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// RemoveOldRunnerProcesses enqueues a signal on the org's signals queue to
// terminate per-process queues (stopping their healthcheck cron emitters) and
// hard-delete all but the most recent runner process per (runner_id, type) for
// the given org. The work runs durably in Temporal so the request returns
// immediately; per-process termination is a retryable activity that survives
// transient Temporal/DB failures.
func (s *service) RemoveOldRunnerProcesses(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()

	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: orgID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org signals queue", zap.String("org_id", orgID), zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find org signals queue"})
		return
	}

	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &removeoldrunnerprocesses.Signal{OrgID: orgID},
	})
	if err != nil {
		s.l.Error("unable to enqueue remove-old-runner-processes signal", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue remove-old-runner-processes signal: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "enqueued",
		"signal_id": resp.ID,
		"queue_id":  queue.ID,
		"message":   "Old runner processes will be cleaned up in the background.",
	})
}
