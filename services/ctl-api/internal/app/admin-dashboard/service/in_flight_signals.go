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

func (s *service) InFlightSignals(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Query("namespace")

	signals, err := s.getInFlightSignals(ctx, namespace)
	if err != nil {
		s.l.Error("failed to get in-flight signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch in-flight signals"})
		return
	}

	namespaces, err := s.getDistinctNamespaces(ctx)
	if err != nil {
		s.l.Error("failed to get namespaces", zap.Error(err))
	}

	component := views.InFlightSignals(signals, namespace, namespaces)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) InFlightSignalsTable(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Query("namespace")

	signals, err := s.getInFlightSignals(ctx, namespace)
	if err != nil {
		s.l.Error("failed to get in-flight signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch in-flight signals"})
		return
	}

	component := views.InFlightSignalsTable(signals, namespace)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getInFlightSignals(ctx context.Context, namespace string) ([]app.QueueSignal, error) {
	var signals []app.QueueSignal

	query := s.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("status->>'status' IN ('executing', 'in-progress', 'active')").
		Order("updated_at DESC").
		Limit(200)

	if namespace != "" {
		query = query.Where("workflow->>'namespace' = ?", namespace)
	}

	if res := query.Find(&signals); res.Error != nil {
		return nil, fmt.Errorf("unable to get in-flight signals: %w", res.Error)
	}

	return signals, nil
}
