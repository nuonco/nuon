package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	queuemigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/queue_migration"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						AdminMigrateOrgQueues
// @Summary				migrate org queues
// @Description			Create all missing queues for this org and enable the queues feature flag
// @Param					org_id	path	string	true	"org ID"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-migrate-queues [POST]
func (s *service) AdminMigrateOrgQueues(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Ensure the org-signals queue exists (it's needed to enqueue the migration signal).
	if err := s.helpers.EnsureOrgQueue(ctx, org.ID); err != nil {
		s.l.Error("unable to ensure org queue", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to ensure org queue: " + err.Error()})
		return
	}

	// Get the org-signals queue ID.
	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: org.ID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org queue", zap.String("org_id", org.ID), zap.Error(res.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find org queue"})
		return
	}

	// Enqueue the queue_migration signal.
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &queuemigration.Signal{OrgID: org.ID},
	}); err != nil {
		s.l.Error("unable to enqueue migration signal", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to enqueue migration signal: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, true)
}
