package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) InstallDetail(c *gin.Context) {
	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.l.Error("failed to fetch install", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	component := views.InstallDetail(install)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// getInstall fetches an install by ID with necessary preloads
func (s *service) getInstall(c *gin.Context) (*app.Install, error) {
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
