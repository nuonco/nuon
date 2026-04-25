package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	queuemigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/queue_migration"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) MigrateOrgQueues(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("unable to get org", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Ensure the org-signals queue exists first (it's needed to enqueue the migration signal)
	if err := s.orgsHelpers.EnsureOrgQueue(ctx, org.ID); err != nil {
		s.l.Error("unable to ensure org queue", zap.String("org_id", org.ID), zap.Error(err))
		component := views.MigrateOrgQueuesToast(false, "Failed to ensure org queue: "+err.Error())
		c.Header("Content-Type", "text/html; charset=utf-8")
		templ.Handler(component).ServeHTTP(c.Writer, c.Request)
		return
	}

	// Get the org-signals queue ID
	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: org.ID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org queue", zap.String("org_id", org.ID), zap.Error(res.Error))
		component := views.MigrateOrgQueuesToast(false, "Failed to find org queue")
		c.Header("Content-Type", "text/html; charset=utf-8")
		templ.Handler(component).ServeHTTP(c.Writer, c.Request)
		return
	}

	// Enqueue the queue_migration signal
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &queuemigration.Signal{OrgID: org.ID},
	}); err != nil {
		s.l.Error("unable to enqueue migration signal", zap.String("org_id", org.ID), zap.Error(err))
		component := views.MigrateOrgQueuesToast(false, "Failed to enqueue migration signal: "+err.Error())
		c.Header("Content-Type", "text/html; charset=utf-8")
		templ.Handler(component).ServeHTTP(c.Writer, c.Request)
		return
	}

	component := views.MigrateOrgQueuesToast(true, "Queue migration started for "+org.Name)
	c.Header("Content-Type", "text/html; charset=utf-8")
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
