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

func (s *service) Installs(c *gin.Context) {
	ctx := c.Request.Context()

	installs, err := s.getInstalls(ctx)
	if err != nil {
		s.l.Error("failed to get installs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	component := views.Installs(installs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getInstalls(ctx context.Context) ([]*app.Install, error) {
	var installs []*app.Install

	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("App").
		Preload("RunnerGroup").
		Order("created_at desc").
		Limit(100).
		Find(&installs)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, nil
}
