package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// OrgStatus returns just the status badge for htmx polling
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

	// Return just the status badge component
	component := views.OrgStatusBadge(org.ID, org.Status, org.StatusDescription)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
