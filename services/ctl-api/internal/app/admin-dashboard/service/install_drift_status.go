package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) InstallDriftStatus(c *gin.Context) {
	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.l.Error("failed to fetch install for drift status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	component := views.InstallDriftStatusBadge(install)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
