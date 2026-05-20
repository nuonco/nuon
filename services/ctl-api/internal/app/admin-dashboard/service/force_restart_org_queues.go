package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) ForceRestartOrgQueues(c *gin.Context) {
	orgID := c.Param("id")
	ctx := cctx.SetOrgIDContext(c.Request.Context(), orgID)

	var queues []app.Queue
	if res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Find(&queues); res.Error != nil {
		s.l.Error("failed to list org queues", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list queues"})
		return
	}

	restarted := 0
	for _, q := range queues {
		if err := s.queueClient.ForceRestart(ctx, q.ID); err != nil {
			s.l.Warn("force-restart-org-queues: failed to restart queue",
				zap.String("queue_id", q.ID),
				zap.String("org_id", orgID),
				zap.Error(err))
			continue
		}
		restarted++
	}

	c.JSON(http.StatusOK, gin.H{"status": "force-restarted", "queues_restarted": restarted})
}
