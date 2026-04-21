package service

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// AddInstallLabel handles adding a key:value label to an install.
func (s *service) AddInstallLabel(c *gin.Context) {
	ctx := c.Request.Context()
	installID := c.Param("id")

	if err := c.Request.ParseForm(); err != nil {
		s.l.Error("failed to parse form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	key := strings.TrimSpace(c.Request.FormValue("label_key"))
	value := strings.TrimSpace(c.Request.FormValue("label_value"))

	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Label key is required"})
		return
	}

	var install app.Install
	if err := s.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		s.l.Error("failed to get install", zap.String("install_id", installID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Install not found"})
		return
	}

	install.Labels.Merge(labels.Labels{key: value})

	if err := s.db.WithContext(ctx).Model(&install).Select("labels").Updates(&install).Error; err != nil {
		s.l.Error("failed to update install labels", zap.String("install_id", installID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update labels"})
		return
	}

	component := views.InstallLabelsSection(&install)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// RemoveInstallLabel handles removing a label key from an install.
func (s *service) RemoveInstallLabel(c *gin.Context) {
	ctx := c.Request.Context()
	installID := c.Param("id")
	key := c.Param("key")

	if installID == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Install ID and label key are required"})
		return
	}

	var install app.Install
	if err := s.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		s.l.Error("failed to get install", zap.String("install_id", installID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Install not found"})
		return
	}

	install.Labels.RemoveKeys([]string{key})

	if err := s.db.WithContext(ctx).Model(&install).Select("labels").Updates(&install).Error; err != nil {
		s.l.Error("failed to update install labels", zap.String("install_id", installID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove label"})
		return
	}

	component := views.InstallLabelsSection(&install)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
