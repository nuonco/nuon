package service

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const accountInstallsPerPage = 8

func (s *service) AccountDetail(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")
	page := getPageFromQuery(c)

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

	// Get installs created by this account with pagination
	installs, installsTotalPages, err := s.getInstallsForAccount(ctx, accountID, page)
	if err != nil {
		s.l.Error("failed to get installs for account", zap.Error(err), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch installs: %v", err)})
		return
	}

	component := views.AccountDetail(&account, apps, installs, page, installsTotalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getInstallsForAccount(ctx context.Context, accountID string, page int) ([]*app.Install, int, error) {
	var installs []*app.Install
	var totalCount int64

	// Build base query
	query := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Where("created_by_id = ?", accountID)

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count installs: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(accountInstallsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate offset
	offset := (page - 1) * accountInstallsPerPage

	// Get paginated results
	res := query.
		Preload("Org").
		Preload("App").
		Preload("RunnerGroup").
		Order("created_at desc").
		Limit(accountInstallsPerPage).
		Offset(offset).
		Find(&installs)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, totalPages, nil
}
