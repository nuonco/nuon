package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin_dashboard/service/views"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) InstallRunnerStatus(c *gin.Context) {
	ctx := c.Request.Context()

	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.logger.Error("failed to fetch install for runner status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	component := views.InstallRunnerStatusBadge(install)
	err = component.Render(ctx, c.Writer)
	if err != nil {
		s.logger.Error("failed to render runner status badge", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render badge"})
		return
	}
}
