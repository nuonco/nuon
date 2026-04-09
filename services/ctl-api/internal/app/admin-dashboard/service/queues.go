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

const queuesPerPage = 20

func (s *service) Queues(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	ownerID := c.Query("owner_id")
	ownerType := c.Query("owner_type")
	search := c.Query("search")

	queues, totalPages, err := s.getQueues(ctx, ownerID, ownerType, search, page)
	if err != nil {
		s.l.Error("failed to get queues", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queues"})
		return
	}

	component := views.Queues(queues, ownerID, ownerType, search, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) QueuesTable(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)
	ownerID := c.Query("owner_id")
	ownerType := c.Query("owner_type")
	search := c.Query("search")

	queues, totalPages, err := s.getQueues(ctx, ownerID, ownerType, search, page)
	if err != nil {
		s.l.Error("failed to get queues", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queues"})
		return
	}

	component := views.QueuesTable(queues, ownerID, ownerType, search, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getQueues(ctx context.Context, ownerID, ownerType, search string, page int) ([]app.Queue, int, error) {
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
