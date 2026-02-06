package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// InstallsTable returns just the installs table for htmx polling
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

	// Return just the table component
	component := views.InstallsTable(orgID, installs, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
