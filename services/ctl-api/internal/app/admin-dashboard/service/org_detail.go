package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) OrgDetail(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Org ID is required"})
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get org", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	installs, err := s.getInstallsForOrg(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get installs", zap.String("org_id", orgID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	component := views.OrgDetail(org, installs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	var org app.Org

	res := s.db.WithContext(ctx).
		Where("id = ?", orgID).
		First(&org)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}

func (s *service) getInstallsForOrg(ctx context.Context, orgID string) ([]*app.Install, error) {
	var installs []*app.Install

	res := s.db.WithContext(ctx).
		Preload("App").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Where("org_id = ?", orgID).
		Order("created_at desc").
		Limit(100).
		Find(&installs)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, nil
}
