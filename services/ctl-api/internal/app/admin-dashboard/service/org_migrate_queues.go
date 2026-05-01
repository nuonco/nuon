package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	queuemigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/queue_migration"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) MigrateOrgQueues(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	ctx = cctx.SetOrgIDContext(ctx, orgID)

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("unable to get org", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Ensure the org-signals queue exists first (it's needed to enqueue the migration signal)
	if err := s.orgsHelpers.EnsureOrgQueue(ctx, org.ID); err != nil {
		s.l.Error("unable to ensure org queue", zap.String("org_id", org.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to ensure org queue: " + err.Error(),
		})
		return
	}

	// Get the org-signals queue ID
	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: org.ID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org queue", zap.String("org_id", org.ID), zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to find org queue",
		})
		return
	}

	// Enqueue the queue_migration signal
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &queuemigration.Signal{OrgID: org.ID},
	}); err != nil {
		s.l.Error("unable to enqueue migration signal", zap.String("org_id", org.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to enqueue migration signal: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Queue migration started for " + org.Name,
	})
}
