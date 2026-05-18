package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	clearorgqueues "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/clear_org_queues"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) ClearOrgQueues(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()

	// Find the org's signals queue.
	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: orgID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org signals queue", zap.String("org_id", orgID), zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find org signals queue"})
		return
	}

	// Enqueue the clear-org-queues signal.
	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &clearorgqueues.Signal{OrgID: orgID},
	})
	if err != nil {
		s.l.Error("unable to enqueue clear-org-queues signal", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue clear-org-queues signal: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "enqueued",
		"signal_id": resp.ID,
		"queue_id":  queue.ID,
		"message":   "Clear org queues signal enqueued. Queues will be cleared in the background.",
	})
}
