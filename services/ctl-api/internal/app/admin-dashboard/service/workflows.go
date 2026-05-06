package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const workflowsPerPage = 20

func (s *service) Workflows(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	sort := c.DefaultQuery("sort", "newest")
	typeFilter := c.Query("type")
	page := getPageFromQuery(c)

	workflows, totalPages, err := s.getWorkflows(ctx, search, sort, typeFilter, page)
	if err != nil {
		s.l.Error("failed to fetch workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows":   workflows,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) WorkflowsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	sort := c.DefaultQuery("sort", "newest")
	typeFilter := c.Query("type")
	page := getPageFromQuery(c)

	workflows, totalPages, err := s.getWorkflows(ctx, search, sort, typeFilter, page)
	if err != nil {
		s.l.Error("failed to fetch workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows":   workflows,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) getWorkflows(ctx context.Context, search, sort, typeFilter string, page int) ([]*app.Workflow, int, error) {
	query := s.readDB().WithContext(ctx).
		Model(&app.Workflow{}).
		Preload("CreatedBy")

	// Search by workflow ID or owner ID
	if search != "" {
		search = strings.TrimSpace(search)
		if strings.HasPrefix(search, "inw") {
			query = query.Where("id = ?", search)
		} else {
			query = query.Where("owner_id = ?", search)
		}
	}

	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}

	// Count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count workflows: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(workflowsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Sort
	switch sort {
	case "oldest":
		query = query.Order("created_at ASC")
	default:
		query = query.Order("created_at DESC")
	}

	offset := (page - 1) * workflowsPerPage
	var workflows []*app.Workflow
	if err := query.Offset(offset).Limit(workflowsPerPage).Find(&workflows).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to get workflows: %w", err)
	}

	// Load step counts
	for _, wf := range workflows {
		var count int64
		s.readDB().WithContext(ctx).Model(&app.WorkflowStep{}).Where("install_workflow_id = ?", wf.ID).Count(&count)
		wf.Steps = make([]app.WorkflowStep, count)
	}

	return workflows, totalPages, nil
}
