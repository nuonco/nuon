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

func (s *service) Orgs(c *gin.Context) {
	ctx := c.Request.Context()

	orgs, err := s.getOrgs(ctx)
	if err != nil {
		s.l.Error("failed to get orgs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	component := views.Orgs(orgs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getOrgs(ctx context.Context) ([]*app.Org, error) {
	var orgs []*app.Org

	res := s.db.WithContext(ctx).
		Order("created_at desc").
		Limit(100). // Reasonable limit for admin dashboard
		Find(&orgs)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	return orgs, nil
}
