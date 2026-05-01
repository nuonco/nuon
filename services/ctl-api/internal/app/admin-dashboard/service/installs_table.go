package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InstallsTable returns the installs table data for an org as JSON
func (s *service) InstallsTable(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	page := getPageFromQuery(c)

	installs, totalPages, err := s.getInstallsForOrg(ctx, orgID, page)
	if err != nil {
		s.l.Error("failed to get installs for table", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"org_id":      orgID,
		"installs":    installs,
		"page":        page,
		"total_pages": totalPages,
	})
}
