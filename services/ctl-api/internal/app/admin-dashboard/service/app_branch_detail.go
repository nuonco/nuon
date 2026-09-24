package service

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	appBranchRunsPerPage      = 10
	appBranchWorkflowsPerPage = 10
)

func (s *service) AppBranchDetail(c *gin.Context) {
	ctx := c.Request.Context()

	branch, err := s.getAppBranch(ctx, c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "app branch not found"})
			return
		}
		s.l.Error("failed to fetch app branch", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch app branch"})
		return
	}

	queues, err := s.getAppBranchQueues(ctx, branch.ID)
	if err != nil {
		s.l.Error("failed to fetch app branch queues", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch app branch queues"})
		return
	}

	runs, runsTotalPages, err := s.getAppBranchRuns(ctx, branch.ID, 1)
	if err != nil {
		s.l.Error("failed to fetch app branch runs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch app branch runs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app_branch":       branch,
		"org_name":         branch.Org.Name,
		"app_name":         branch.App.Name,
		"created_by":       branch.CreatedBy,
		"queues":           queues,
		"runs":             runs,
		"runs_total_pages": runsTotalPages,
		"app_url":          s.cfg.AppURL,
	})
}

func (s *service) AppBranchRunsTable(c *gin.Context) {
	ctx := c.Request.Context()
	branchID := c.Param("id")
	page := getPageFromQuery(c)

	runs, totalPages, err := s.getAppBranchRuns(ctx, branchID, page)
	if err != nil {
		s.l.Error("failed to fetch app branch runs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch app branch runs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app_branch_id": branchID,
		"runs":          runs,
		"page":          page,
		"total_pages":   totalPages,
	})
}

func (s *service) AppBranchWorkflowsTable(c *gin.Context) {
	ctx := c.Request.Context()
	branchID := c.Param("id")
	page := getPageFromQuery(c)

	workflows, totalPages, err := s.getAppBranchWorkflows(ctx, branchID, page)
	if err != nil {
		s.l.Error("failed to fetch app branch workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch app branch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"app_branch_id": branchID,
		"workflows":     workflows,
		"page":          page,
		"total_pages":   totalPages,
	})
}

func (s *service) getAppBranch(ctx context.Context, branchID string) (*app.AppBranch, error) {
	if branchID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var branch app.AppBranch
	if err := s.readDB().WithContext(ctx).
		Unscoped().
		Preload("Org").
		Preload("App").
		Preload("CreatedBy").
		Where(&app.AppBranch{ID: branchID}).
		First(&branch).Error; err != nil {
		return nil, err
	}

	var workflowCount int64
	if err := s.readDB().WithContext(ctx).
		Model(&app.Workflow{}).
		Where(&app.Workflow{OwnerID: branch.ID, OwnerType: appBranchOwnerType}).
		Count(&workflowCount).Error; err != nil {
		return nil, fmt.Errorf("unable to count app branch workflows: %w", err)
	}
	branch.WorkflowCount = int(workflowCount)

	return &branch, nil
}

func (s *service) getAppBranchQueues(ctx context.Context, branchID string) ([]app.Queue, error) {
	var queues []app.Queue
	if err := s.readDB().WithContext(ctx).
		Preload("Emitters").
		Where(&app.Queue{OwnerID: branchID, OwnerType: appBranchOwnerType}).
		Order("created_at DESC").
		Find(&queues).Error; err != nil {
		return nil, fmt.Errorf("unable to get app branch queues: %w", err)
	}

	return queues, nil
}

func (s *service) getAppBranchRuns(ctx context.Context, branchID string, page int) ([]*app.AppBranchRun, int, error) {
	if branchID == "" {
		return nil, 0, gorm.ErrRecordNotFound
	}

	var totalCount int64
	if err := s.readDB().WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(&app.AppBranchRun{AppBranchID: branchID}).
		Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count app branch runs: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(appBranchRunsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * appBranchRunsPerPage
	var runs []*app.AppBranchRun
	if err := s.readDB().WithContext(ctx).
		Preload("CreatedBy").
		Preload("VCSConnectionCommit").
		Where(&app.AppBranchRun{AppBranchID: branchID}).
		Order("created_at DESC").
		Offset(offset).
		Limit(appBranchRunsPerPage).
		Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to get app branch runs: %w", err)
	}

	return runs, totalPages, nil
}

func (s *service) getAppBranchWorkflows(ctx context.Context, branchID string, page int) ([]*app.Workflow, int, error) {
	if branchID == "" {
		return nil, 0, gorm.ErrRecordNotFound
	}

	owner := app.Workflow{OwnerID: branchID, OwnerType: appBranchOwnerType}

	var totalCount int64
	if err := s.readDB().WithContext(ctx).
		Model(&app.Workflow{}).
		Where(&owner).
		Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count app branch workflows: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(appBranchWorkflowsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * appBranchWorkflowsPerPage
	var workflows []*app.Workflow
	if err := s.readDB().WithContext(ctx).
		Preload("CreatedBy").
		Where(&owner).
		Order("created_at DESC").
		Offset(offset).
		Limit(appBranchWorkflowsPerPage).
		Find(&workflows).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to get app branch workflows: %w", err)
	}

	return workflows, totalPages, nil
}
