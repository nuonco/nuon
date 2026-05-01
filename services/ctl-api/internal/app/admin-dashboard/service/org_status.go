package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrgStatus returns the org status as JSON
func (s *service) OrgStatus(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org status", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":             org.Status,
		"status_description": org.StatusDescription,
	})
}
