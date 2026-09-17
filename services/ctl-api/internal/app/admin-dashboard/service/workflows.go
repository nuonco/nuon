package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const workflowsPerPage = 20

var workflowsPerPageOptions = []int{20, 50}

func getPerPageFromQuery(c *gin.Context) int {
	perPage, err := strconv.Atoi(c.Query("per_page"))
	if err != nil {
		return workflowsPerPage
	}
	for _, allowed := range workflowsPerPageOptions {
		if perPage == allowed {
			return perPage
		}
	}
	return workflowsPerPage
}

type workflowFilter struct {
	Search        string
	Type          string
	Status        string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

func (s *service) Workflows(c *gin.Context) {
	s.respondWorkflows(c)
}

func (s *service) WorkflowsTable(c *gin.Context) {
	s.respondWorkflows(c)
}

func (s *service) respondWorkflows(c *gin.Context) {
	ctx := c.Request.Context()
	sort := c.DefaultQuery("sort", "newest")
	page := getPageFromQuery(c)
	perPage := getPerPageFromQuery(c)

	filter, err := workflowFilterFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflows, total, totalPages, err := s.getWorkflows(ctx, filter, sort, page, perPage)
	if err != nil {
		s.l.Error("failed to fetch workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows":   workflows,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	})
}

func workflowFilterFromQuery(c *gin.Context) (workflowFilter, error) {
	filter := workflowFilter{
		Search: strings.TrimSpace(c.Query("search")),
		Type:   c.Query("type"),
		Status: c.Query("status"),
	}

	after, err := parseTimeQuery(c.Query("created_after"))
	if err != nil {
		return filter, fmt.Errorf("invalid created_after: %w", err)
	}
	filter.CreatedAfter = after

	before, err := parseTimeQuery(c.Query("created_before"))
	if err != nil {
		return filter, fmt.Errorf("invalid created_before: %w", err)
	}
	filter.CreatedBefore = before

	return filter, nil
}

func parseTimeQuery(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (s *service) workflowQuery(ctx context.Context, filter workflowFilter) *gorm.DB {
	query := s.readDB().WithContext(ctx).Model(&app.Workflow{})

	if filter.Search != "" {
		if strings.HasPrefix(filter.Search, "inw") {
			query = query.Where("id = ?", filter.Search)
		} else {
			query = query.Where("owner_id = ?", filter.Search)
		}
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.Status != "" {
		query = query.Where("status->>'status' = ?", filter.Status)
	}

	if filter.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filter.CreatedAfter)
	}

	if filter.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filter.CreatedBefore)
	}

	return query
}

func (s *service) getWorkflows(ctx context.Context, filter workflowFilter, sort string, page, perPage int) ([]*app.Workflow, int64, int, error) {
	query := s.workflowQuery(ctx, filter)

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("unable to count workflows: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	switch sort {
	case "oldest":
		query = query.Order("created_at ASC")
	default:
		query = query.Order("created_at DESC")
	}

	offset := (page - 1) * perPage
	var workflows []*app.Workflow
	if err := query.Preload("CreatedBy").Offset(offset).Limit(perPage).Find(&workflows).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("unable to get workflows: %w", err)
	}

	for _, wf := range workflows {
		var count int64
		s.readDB().WithContext(ctx).Model(&app.WorkflowStep{}).Where("install_workflow_id = ?", wf.ID).Count(&count)
		wf.Steps = make([]app.WorkflowStep, count)
	}

	return workflows, totalCount, totalPages, nil
}
