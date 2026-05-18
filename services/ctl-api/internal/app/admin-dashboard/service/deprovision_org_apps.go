package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// DeprovisionOrgApps marks all apps in an org as delete-queued.
func (s *service) DeprovisionOrgApps(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()

	var apps []app.App
	if res := s.db.WithContext(ctx).
		Where(app.App{OrgID: orgID}).
		Find(&apps); res.Error != nil {
		s.l.Error("failed to list org apps", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list apps"})
		return
	}

	updated := 0
	var errs []string
	for _, a := range apps {
		if a.Status == app.AppStatusDeleteQueued || a.Status == app.AppStatusDeprovisioning {
			continue
		}

		res := s.db.WithContext(ctx).
			Model(&app.App{ID: a.ID}).
			Updates(app.App{
				Status:            app.AppStatusDeleteQueued,
				StatusDescription: "delete queued via admin cleanup",
			})
		if res.Error != nil {
			s.l.Warn("failed to mark app for deletion",
				zap.String("app_id", a.ID), zap.Error(res.Error))
			errs = append(errs, a.ID)
			continue
		}
		updated++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "deprovisioning",
		"apps_updated": updated,
		"apps_total":   len(apps),
		"errors":       errs,
	})
}
