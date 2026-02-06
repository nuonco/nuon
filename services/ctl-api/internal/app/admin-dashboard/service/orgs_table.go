package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// OrgsTable returns just the orgs table for htmx polling
func (s *service) OrgsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	page := getPageFromQuery(c)

	orgs, totalPages, err := s.getOrgs(ctx, search, page)
	if err != nil {
		s.l.Error("failed to get orgs for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	// Return just the table component
	component := views.OrgsTable(orgs, page, totalPages, search)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
