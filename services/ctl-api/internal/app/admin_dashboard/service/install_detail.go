package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin_dashboard/service/views"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) InstallDetail(c *gin.Context) {
	ctx := c.Request.Context()

	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.logger.Error("failed to fetch install", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	component := views.InstallDetail(install)
	err = component.Render(ctx, c.Writer)
	if err != nil {
		s.logger.Error("failed to render install detail", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render page"})
		return
	}
}

// getInstall fetches an install by ID with necessary preloads
func (s *Service) getInstall(c *gin.Context) (*app.Install, error) {
	installID := c.Param("id")
	if installID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var install app.Install
	err := s.db.
		Preload("Org").
		Preload("App").
		Preload("AppConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AppRunnerConfig").
		Where("id = ?", installID).
		First(&install).Error

	if err != nil {
		return nil, err
	}

	return &install, nil
}
