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

const queuesPerPage = 20

func (s *service) Queues(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	ownerID := c.Query("owner_id")
	ownerType := c.Query("owner_type")
	search := c.Query("search")
	queueName := c.Query("queue_name")
	namespace := c.Query("namespace")
	redirect := c.Query("redirect")

	queues, totalPages, err := s.getQueues(ctx, ownerID, ownerType, search, queueName, namespace, page)
	if err != nil {
		s.l.Error("failed to get queues", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queues"})
		return
	}

	if redirect == "true" && len(queues) == 1 {
		c.JSON(http.StatusOK, gin.H{"redirect": fmt.Sprintf("/queues/%s", queues[0].ID)})
		return
	}

	namespaces, err := s.getDistinctQueueNamespaces(ctx)
	if err != nil {
		s.l.Error("failed to get queue namespaces", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"queues":      queues,
		"namespaces":  namespaces,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) QueuesTable(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	ownerID := c.Query("owner_id")
	ownerType := c.Query("owner_type")
	search := c.Query("search")
	queueName := c.Query("queue_name")
	namespace := c.Query("namespace")

	queues, totalPages, err := s.getQueues(ctx, ownerID, ownerType, search, queueName, namespace, page)
	if err != nil {
		s.l.Error("failed to get queues", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queues"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queues":      queues,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) getQueues(ctx context.Context, ownerID, ownerType, search, queueName, namespace string, page int) ([]app.Queue, int, error) {
	var queues []app.Queue
	var totalCount int64

	query := s.db.WithContext(ctx).Model(&app.Queue{})

	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	if ownerType != "" {
		query = query.Where("owner_type = ?", ownerType)
	}
	if search != "" {
		query = query.Where("id = ? OR owner_id = ? OR name ILIKE ?", search, search, "%"+search+"%")
	}
	if queueName != "" {
		query = query.Where("name = ?", queueName)
	}
	if namespace != "" {
		query = query.Where("workflow->>'namespace' = ?", namespace)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count queues: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(queuesPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * queuesPerPage

	res := query.
		Preload("Emitters").
		Order("created_at desc").
		Limit(queuesPerPage).
		Offset(offset).
		Find(&queues)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get queues: %w", res.Error)
	}

	return queues, totalPages, nil
}

func (s *service) getDistinctQueueNamespaces(ctx context.Context) ([]string, error) {
	var namespaces []string
	res := s.db.WithContext(ctx).
		Model(&app.Queue{}).
		Select("DISTINCT workflow->>'namespace'").
		Where("workflow->>'namespace' != ''").
		Pluck("workflow->>'namespace'", &namespaces)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get distinct queue namespaces: %w", res.Error)
	}
	sort.Strings(namespaces)
	return namespaces, nil
}
