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

func (s *service) Installs(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")

	installs, err := s.getInstalls(ctx, search)
	if err != nil {
		s.l.Error("failed to get installs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	component := views.Installs(installs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getInstalls(ctx context.Context, search string) ([]*app.Install, error) {
	var installs []*app.Install

	query := s.db.WithContext(ctx).
		Preload("Org").
		Preload("App").
		Preload("RunnerGroup")

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
		Limit(100).
		Find(&installs)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, nil
}
