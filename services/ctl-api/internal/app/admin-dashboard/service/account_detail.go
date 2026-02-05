package service

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) AccountDetail(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")

	var account app.Account
	res := s.db.WithContext(ctx).
		Preload("Roles.Org").
		Where("id = ?", accountID).
		First(&account)

	if res.Error != nil {
		s.l.Error("failed to get account", zap.Error(res.Error), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch account: %v", res.Error)})
		return
	}

	// Get apps created by this account
	var apps []*app.App
	appRes := s.db.WithContext(ctx).
		Preload("Org").
		Where("created_by_id = ?", accountID).
		Order("created_at desc").
		Limit(100).
		Find(&apps)

	if appRes.Error != nil {
		s.l.Error("failed to get apps for account", zap.Error(appRes.Error), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch apps: %v", appRes.Error)})
		return
	}

	// Get installs created by this account
	var installs []*app.Install
	installRes := s.db.WithContext(ctx).
		Preload("Org").
		Preload("App").
		Preload("RunnerGroup").
		Where("created_by_id = ?", accountID).
		Order("created_at desc").
		Limit(100).
		Find(&installs)

	if installRes.Error != nil {
		s.l.Error("failed to get installs for account", zap.Error(installRes.Error), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch installs: %v", installRes.Error)})
		return
	}

	component := views.AccountDetail(&account, apps, installs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
