package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// InstallsTableGlobal returns just the installs table for htmx polling
func (s *service) InstallsTableGlobal(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")

	installs, err := s.getInstalls(ctx, search)
	if err != nil {
		s.l.Error("failed to get installs for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	component := views.InstallsTableGlobal(installs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
