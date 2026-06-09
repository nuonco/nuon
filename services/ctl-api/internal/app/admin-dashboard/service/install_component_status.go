package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *service) InstallComponentStatus(c *gin.Context) {
	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.l.Error("failed to fetch install for component status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"install_id":      install.ID,
		"lifecycle_phase": install.LifecyclePhase.Phase,
	})
}
