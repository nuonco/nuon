package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const installActivityPerPage = 10

func (s *service) InstallDetail(c *gin.Context) {
	ctx := c.Request.Context()
	page := getPageFromQuery(c)

	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	install, err := s.getInstall(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "install not found"})
			return
		}
		s.l.Error("failed to fetch install", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch install"})
		return
	}

	activeDeployments, err := s.getActiveDeployments(c, install.ID)
	if err != nil {
		s.l.Error("failed to fetch active deployments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch active deployments"})
		return
	}

	// Parse entity type filters
	var entityTypes []string
	typeFilter := c.Query("entity_types")
	if typeFilter != "" {
		for _, t := range strings.Split(typeFilter, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				entityTypes = append(entityTypes, trimmed)
			}
		}
	}
	// If we ended up with an empty array after parsing, set to nil so defaults work
	if len(entityTypes) == 0 {
		entityTypes = nil
	}

	// Fetch activity logs
	activityLogs, activityTotalPages, err := s.getActivityForInstall(
		ctx, install.ID, startDate, endDate, page, entityTypes,
	)
	if err != nil {
		s.l.Warn("failed to get activity logs", zap.Error(err))
		activityLogs = []*AuditLogEntry{}
		activityTotalPages = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"install":              install,
		"active_deployments":   activeDeployments,
		"activity_logs":        activityLogs,
		"start_date":           startDate,
		"end_date":             endDate,
		"app_url":              s.cfg.AppURL,
		"page":                 page,
		"activity_total_pages": activityTotalPages,
	})
}

// InstallActiveDeploymentsTable handles the polling endpoint for active deployments
func (s *service) InstallActiveDeploymentsTable(c *gin.Context) {
	installID := c.Param("id")
	if installID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "install ID required"})
		return
	}

	deployments, err := s.getActiveDeployments(c, installID)
	if err != nil {
		s.l.Error("failed to fetch active deployments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch deployments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"install_id":  installID,
		"deployments": deployments,
	})
}

// InstallActivityTable handles the polling endpoint for install activity
func (s *service) InstallActivityTable(c *gin.Context) {
	ctx := c.Request.Context()
	installID := c.Param("id")
	page := getPageFromQuery(c)

	if installID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "install ID required"})
		return
	}

	// Parse date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	// Parse entity type filters
	var entityTypes []string
	typeFilter := c.Query("entity_types")
	if typeFilter != "" {
		for _, t := range strings.Split(typeFilter, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				entityTypes = append(entityTypes, trimmed)
			}
		}
	}
	// If we ended up with an empty array after parsing, set to nil so defaults work
	if len(entityTypes) == 0 {
		entityTypes = nil
	}

	activityLogs, activityTotalPages, err := s.getActivityForInstall(
		ctx, installID, startDate, endDate, page, entityTypes,
	)
	if err != nil {
		s.l.Error("failed to fetch activity logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"install_id":    installID,
		"activity_logs": activityLogs,
		"start_date":    startDate,
		"end_date":      endDate,
		"page":          page,
		"total_pages":   activityTotalPages,
	})
}

const installWorkflowsPerPage = 10

// InstallWorkflowsTable handles the endpoint for install workflows
func (s *service) InstallWorkflowsTable(c *gin.Context) {
	ctx := c.Request.Context()
	installID := c.Param("id")
	page := getPageFromQuery(c)

	if installID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "install ID required"})
		return
	}

	workflows, totalPages, err := s.getWorkflowsForInstall(ctx, installID, page)
	if err != nil {
		s.l.Error("failed to fetch install workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"install_id":  installID,
		"workflows":   workflows,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) getWorkflowsForInstall(ctx context.Context, installID string, page int) ([]*app.Workflow, int, error) {
	var totalCount int64
	countQuery := s.readDB().WithContext(ctx).
		Model(&app.Workflow{}).
		Where("owner_id = ? AND owner_type = 'installs'", installID)

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count workflows: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(installWorkflowsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * installWorkflowsPerPage
	var workflows []*app.Workflow
	if err := s.readDB().WithContext(ctx).
		Where("owner_id = ? AND owner_type = 'installs'", installID).
		Order("created_at DESC").
		Offset(offset).
		Limit(installWorkflowsPerPage).
		Find(&workflows).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to get workflows: %w", err)
	}

	return workflows, totalPages, nil
}

// getInstall fetches an install by ID with necessary preloads
func (s *service) getInstall(c *gin.Context) (*app.Install, error) {
	installID := c.Param("id")
	if installID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var install app.Install
	err := s.readDB().
		Unscoped().
		Preload("Org").
		Preload("App").
		Preload("AppConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AppRunnerConfig").
		Preload("Queues").
		Where("id = ?", installID).
		First(&install).Error

	if err != nil {
		return nil, err
	}

	return &install, nil
}

// getActiveDeployments fetches active deployments for an install
func (s *service) getActiveDeployments(c *gin.Context, installID string) ([]app.InstallDeploy, error) {
	activeStatuses := []app.InstallDeployStatus{
		app.InstallDeployStatusPlanning,
		app.InstallDeployStatusSyncing,
		app.InstallDeployStatusExecuting,
		app.InstallDeployStatusQueued,
		app.InstallDeployStatusPending,
		app.InstallDeployStatusPendingApproval,
	}

	var deployments []app.InstallDeploy
	err := s.readDB().
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_components.install_id = ?", installID).
		Where("install_deploys.status IN ?", activeStatuses).
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Order("install_deploys.created_at DESC").
		Find(&deployments).Error

	if err != nil {
		return nil, err
	}

	for i := range deployments {
		if deployments[i].InstallComponent.Component.Name != "" {
			deployments[i].ComponentName = deployments[i].InstallComponent.Component.Name
		}
	}

	return deployments, nil
}

// getActivityForInstall fetches runner jobs and workflows for an install
func (s *service) getActivityForInstall(
	ctx context.Context,
	installID string,
	startDate, endDate time.Time,
	page int,
	entityTypes []string,
) ([]*AuditLogEntry, int, error) {
	var entries []*AuditLogEntry

	// Default to runner_job and workflow if no types specified
	// Also handle case where empty string is passed
	if len(entityTypes) == 0 || (len(entityTypes) == 1 && entityTypes[0] == "") {
		entityTypes = []string{"runner_job", "workflow"}
		s.l.Debug("using default entity types for install activity",
			zap.String("install_id", installID),
			zap.Strings("entity_types", entityTypes),
		)
	} else {
		s.l.Debug("using filtered entity types for install activity",
			zap.String("install_id", installID),
			zap.Strings("entity_types", entityTypes),
		)
	}

	// Build queries based on selected entity types
	var queries []string
	var queryParams []interface{}

	for _, entityType := range entityTypes {
		switch entityType {
		case "runner_job":
			// Runner jobs are related to installs through multiple owner types
			// We need to join through install_deploys, install_sandbox_runs, install_components, and install_action_workflow_runs
			queries = append(queries, `
				SELECT DISTINCT
					'runner_job' as entity_type,
					rj.id as entity_id,
					CONCAT(CAST(rj.type AS TEXT), ' - ', CAST(rj.operation AS TEXT)) as entity_name,
					rj.created_at,
					rj.org_id,
					o.name as org_name,
					NULL as app_id,
					NULL as app_name,
					CAST(rj.status AS TEXT) as description
				FROM runner_jobs rj
				LEFT JOIN orgs o ON rj.org_id = o.id
				LEFT JOIN install_deploys id ON rj.owner_id = id.id AND rj.owner_type = 'install_deploys'
				LEFT JOIN install_components ic ON id.install_component_id = ic.id
				LEFT JOIN install_sandbox_runs isr ON rj.owner_id = isr.id AND rj.owner_type = 'install_sandbox_runs'
				LEFT JOIN install_action_workflow_runs iawr ON rj.owner_id = iawr.id AND rj.owner_type = 'install_action_workflow_runs'
				WHERE (
					(ic.install_id = ? AND rj.owner_type = 'install_deploys') OR
					(isr.install_id = ? AND rj.owner_type = 'install_sandbox_runs') OR
					(rj.owner_id IN (SELECT id FROM install_components WHERE install_id = ?) AND rj.owner_type = 'install_components') OR
					(iawr.install_id = ? AND rj.owner_type = 'install_action_workflow_runs')
				) AND rj.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, installID, installID, installID, installID, startDate, endDate)

		case "workflow":
			// Query install workflows (table name is install_workflows)
			queries = append(queries, `
				SELECT
					'workflow' as entity_type,
					iw.id as entity_id,
					CAST(iw.type AS TEXT) as entity_name,
					iw.created_at,
					iw.org_id,
					o.name as org_name,
					NULL as app_id,
					a.name as app_name,
					COALESCE(iw.status->>'status', 'unknown') as description
				FROM install_workflows iw
				LEFT JOIN orgs o ON iw.org_id = o.id
				LEFT JOIN installs i ON iw.owner_id = i.id
				LEFT JOIN apps a ON i.app_id = a.id
				WHERE iw.owner_id = ? AND iw.owner_type = 'installs' AND iw.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, installID, startDate, endDate)
		}
	}

	if len(queries) == 0 {
		return []*AuditLogEntry{}, 1, nil
	}

	// Join queries with UNION ALL
	query := ""
	for i, q := range queries {
		if i > 0 {
			query += " UNION ALL "
		}
		query += q
	}
	query += " ORDER BY created_at DESC"

	// Get total count
	countQuery := `SELECT COUNT(*) FROM (` + query + `) as activity_entries`
	var totalCount int64
	err := s.readDB().WithContext(ctx).Raw(countQuery, queryParams...).Scan(&totalCount).Error
	if err != nil {
		s.l.Error("failed to count activity", zap.Error(err), zap.String("install_id", installID))
		return nil, 0, fmt.Errorf("unable to count activity: %w", err)
	}

	s.l.Debug("activity query results",
		zap.String("install_id", installID),
		zap.Int64("total_count", totalCount),
		zap.Strings("entity_types", entityTypes),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	// Calculate pagination
	totalPages := int(math.Ceil(float64(totalCount) / float64(installActivityPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * installActivityPerPage

	// Execute paginated query
	queryParams = append(queryParams, installActivityPerPage, offset)
	err = s.readDB().WithContext(ctx).Raw(query+` LIMIT ? OFFSET ?`, queryParams...).Scan(&entries).Error
	if err != nil {
		return nil, 0, fmt.Errorf("unable to get activity: %w", err)
	}

	return entries, totalPages, nil
}
