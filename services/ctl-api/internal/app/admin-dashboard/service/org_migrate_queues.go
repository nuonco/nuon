package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	queuemigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/queue_migration"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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

	if err := s.orgsHelpers.EnqueueOrgSignal(ctx, orgshelpers.EnqueueOrgSignalParams{
		OrgID:  org.ID,
		Signal: &queuemigration.Signal{OrgID: org.ID},
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
