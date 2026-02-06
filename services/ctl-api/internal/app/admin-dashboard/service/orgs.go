package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) Orgs(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")

	orgs, err := s.getOrgs(ctx, search)
	if err != nil {
		s.l.Error("failed to get orgs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	component := views.Orgs(orgs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getOrgs(ctx context.Context, search string) ([]*app.Org, error) {
	var orgs []*app.Org

	query := s.db.WithContext(ctx)

	// Apply search filter if provided
	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"name ILIKE ? OR id = ?",
			"%"+search+"%",
			search,
		)
	}

	res := query.
		Order("created_at desc").
		Limit(100). // Reasonable limit for admin dashboard
		Find(&orgs)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	return orgs, nil
}
