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

var allowedSinceValues = map[string]string{
	"15m": "15 minutes",
	"1h":  "1 hour",
	"12h": "12 hours",
	"24h": "24 hours",
}

func (s *service) QueueSignals(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	search := c.Query("search")
	ownerID := c.Query("owner_id")
	orgID := c.Query("org_id")
	signalType := c.Query("signal_type")
	status := c.Query("status")
	namespace := c.Query("namespace")
	enqueued := c.Query("enqueued")
	sortBy := c.Query("sort_by")
	since := c.DefaultQuery("since", "15m")

	signals, totalPages, err := s.getQueueSignals(ctx, search, ownerID, orgID, signalType, status, namespace, enqueued, sortBy, since, page)
	if err != nil {
		s.l.Error("failed to get queue signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queue signals"})
		return
	}

	namespaces, err := s.getDistinctNamespaces(ctx, since)
	if err != nil {
		s.l.Error("failed to get namespaces", zap.Error(err))
	}

	signalTypes, err := s.getDistinctSignalTypes(ctx, namespace, since)
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
	orgID := c.Query("org_id")
	signalType := c.Query("signal_type")
	status := c.Query("status")
	namespace := c.Query("namespace")
	enqueued := c.Query("enqueued")
	sortBy := c.Query("sort_by")
	since := c.DefaultQuery("since", "15m")

	signals, totalPages, err := s.getQueueSignals(ctx, search, ownerID, orgID, signalType, status, namespace, enqueued, sortBy, since, page)
	if err != nil {
		s.l.Error("failed to get queue signals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queue signals"})
		return
	}

	namespaces, err := s.getDistinctNamespaces(ctx, since)
	if err != nil {
		s.l.Warn("failed to get namespaces", zap.Error(err))
	}

	signalTypes, err := s.getDistinctSignalTypes(ctx, namespace, since)
	if err != nil {
		s.l.Warn("failed to get signal types", zap.Error(err))
	}

	orgOptions := s.getOrgOptions(ctx)

	c.JSON(http.StatusOK, gin.H{
		"signals":      signals,
		"page":         page,
		"total_pages":  totalPages,
		"namespaces":   namespaces,
		"signal_types": signalTypes,
		"org_options":  orgOptions,
	})
}

// QueueSignalTypeOptions returns the signal type options filtered by namespace.
func (s *service) QueueSignalTypeOptions(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Query("namespace")
	since := c.DefaultQuery("since", "24h")

	signalTypes, err := s.getDistinctSignalTypes(ctx, namespace, since)
	if err != nil {
		s.l.Error("failed to get signal types", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"signal_types": signalTypes,
	})
}

var allowedSortColumns = map[string]string{
	"created_at":      "created_at desc",
	"updated_at":      "updated_at desc",
	"execution_count": "execution_count desc",
}

func (s *service) getQueueSignals(ctx context.Context, search, ownerID, orgID, signalType, status, namespace, enqueued, sortBy, since string, page int) ([]app.QueueSignal, int, error) {
	var signals []app.QueueSignal
	var totalCount int64

	query := s.readDB().WithContext(ctx).Model(&app.QueueSignal{})

	// Apply time window filter to avoid scanning millions of rows.
	if interval, ok := allowedSinceValues[since]; ok {
		query = query.Where("created_at >= NOW() - INTERVAL '" + interval + "'")
	}

	if search != "" {
		switch {
		case len(search) >= 3 && search[:3] == "qsi":
			query = query.Where("id = ?", search)
		case len(search) >= 3 && search[:3] == "que":
			query = query.Where("queue_id = ?", search)
		default:
			query = query.Where("owner_id = ?", search)
		}
	}
	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	if orgID != "" {
		query = query.Where("org_id = ?", orgID)
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

	orderClause := "created_at desc"
	if col, ok := allowedSortColumns[sortBy]; ok {
		orderClause = col
	}

	res := query.
		Order(orderClause).
		Limit(queueSignalsPerPage).
		Offset(offset).
		Find(&signals)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get queue signals: %w", res.Error)
	}

	return signals, totalPages, nil
}

func (s *service) getDistinctNamespaces(ctx context.Context, since string) ([]string, error) {
	var namespaces []string
	query := s.readDB().WithContext(ctx).
		Model(&app.QueueSignal{}).
		Select("DISTINCT workflow->>'namespace'").
		Where("workflow->>'namespace' != ''")
	if interval, ok := allowedSinceValues[since]; ok {
		query = query.Where("created_at >= NOW() - INTERVAL '" + interval + "'")
	}
	res := query.Pluck("workflow->>'namespace'", &namespaces)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get distinct namespaces: %w", res.Error)
	}
	sort.Strings(namespaces)
	return namespaces, nil
}

func (s *service) getDistinctSignalTypes(ctx context.Context, namespace string, since string) ([]string, error) {
	var types []string
	query := s.readDB().WithContext(ctx).Model(&app.QueueSignal{})
	if interval, ok := allowedSinceValues[since]; ok {
		query = query.Where("created_at >= NOW() - INTERVAL '" + interval + "'")
	}
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
