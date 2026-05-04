package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const queueSignalsPerPage = 100

func (s *service) QueueSignals(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	search := c.Query("search")
	ownerID := c.Query("owner_id")
	signalType := c.Query("signal_type")
	status := c.Query("status")
	namespace := c.Query("namespace")
	enqueued := c.Query("enqueued")

	signals, totalPages, err := s.getQueueSignals(ctx, search, ownerID, signalType, status, namespace, enqueued, page)
	if err != nil {
		s.l.Error("failed to get queue signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queue signals"})
		return
	}

	namespaces, err := s.getDistinctNamespaces(ctx)
	if err != nil {
		s.l.Error("failed to get namespaces", zap.Error(err))
	}

	signalTypes, err := s.getDistinctSignalTypes(ctx, namespace)
	if err != nil {
		s.l.Error("failed to get signal types", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"signals":      signals,
		"namespaces":   namespaces,
		"signal_types": signalTypes,
		"page":         page,
		"total_pages":  totalPages,
	})
}

func (s *service) QueueSignalsGlobalTable(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	search := c.Query("search")
	ownerID := c.Query("owner_id")
	signalType := c.Query("signal_type")
	status := c.Query("status")
	namespace := c.Query("namespace")
	enqueued := c.Query("enqueued")

	signals, totalPages, err := s.getQueueSignals(ctx, search, ownerID, signalType, status, namespace, enqueued, page)
	if err != nil {
		s.l.Error("failed to get queue signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queue signals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signals":     signals,
		"page":        page,
		"total_pages": totalPages,
	})
}

// QueueSignalTypeOptions returns the signal type options filtered by namespace.
func (s *service) QueueSignalTypeOptions(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Query("namespace")

	signalTypes, err := s.getDistinctSignalTypes(ctx, namespace)
	if err != nil {
		s.l.Error("failed to get signal types", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"signal_types": signalTypes,
	})
}

func (s *service) getQueueSignals(ctx context.Context, search, ownerID, signalType, status, namespace, enqueued string, page int) ([]app.QueueSignal, int, error) {
	var signals []app.QueueSignal
	var totalCount int64

	query := s.db.WithContext(ctx).Model(&app.QueueSignal{})

	if search != "" {
		query = query.Where("id = ? OR owner_id = ? OR queue_id = ?", search, search, search)
	}
	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	if signalType != "" {
		query = query.Where("type = ?", signalType)
	}
	if status != "" {
		query = query.Where("status->>'status' = ?", status)
	}
	if namespace != "" {
		query = query.Where("workflow->>'namespace' = ?", namespace)
	}
	if enqueued == "true" {
		query = query.Where("enqueued = ?", true)
	} else if enqueued == "false" {
		query = query.Where("enqueued = ?", false)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count queue signals: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(queueSignalsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * queueSignalsPerPage

	res := query.
		Order("created_at desc").
		Limit(queueSignalsPerPage).
		Offset(offset).
		Find(&signals)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get queue signals: %w", res.Error)
	}

	return signals, totalPages, nil
}

func (s *service) getDistinctNamespaces(ctx context.Context) ([]string, error) {
	var namespaces []string
	res := s.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Select("DISTINCT workflow->>'namespace'").
		Where("workflow->>'namespace' != ''").
		Pluck("workflow->>'namespace'", &namespaces)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get distinct namespaces: %w", res.Error)
	}
	sort.Strings(namespaces)
	return namespaces, nil
}

func (s *service) getDistinctSignalTypes(ctx context.Context, namespace string) ([]string, error) {
	var types []string
	query := s.db.WithContext(ctx).Model(&app.QueueSignal{})
	if namespace != "" {
		query = query.Where("workflow->>'namespace' = ?", namespace)
	}
	res := query.
		Select("DISTINCT type").
		Where("type != ''").
		Pluck("type", &types)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get distinct signal types: %w", res.Error)
	}
	sort.Strings(types)
	return types, nil
}
