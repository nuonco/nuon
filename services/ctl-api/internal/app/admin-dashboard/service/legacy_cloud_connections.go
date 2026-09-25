package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) LegacyCloudConnections(c *gin.Context) {
	query := s.readDB().WithContext(c.Request.Context()).Where(app.CloudConnection{AuthMode: app.CloudConnectionAuthModeLegacy})
	if orgID := c.Query("org_id"); orgID != "" {
		query = query.Where(app.CloudConnection{OrgID: orgID})
	}
	var connections []app.CloudConnection
	if err := query.Order("org_id, name").Find(&connections).Error; err != nil {
		s.l.Error("failed to list legacy cloud connections", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch legacy cloud connections"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"connections": connections})
}
