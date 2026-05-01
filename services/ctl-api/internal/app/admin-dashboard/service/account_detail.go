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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const accountInstallsPerPage = 8
const accountAuditLogsPerPage = 8

// AuditLogEntry represents a single audit log entry for account or install activity.
type AuditLogEntry struct {
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	EntityName  string    `json:"entity_name"`
	CreatedAt   time.Time `json:"created_at"`
	OrgID       *string   `json:"org_id"`
	OrgName     *string   `json:"org_name"`
	AppID       *string   `json:"app_id"`
	AppName     *string   `json:"app_name"`
	Description *string   `json:"description"`
}

func (s *service) AccountDetail(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")
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
			// Set to end of day (23:59:59) to include all entries from that day
			endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	var account app.Account
	res := s.db.WithContext(ctx).
		Preload("Roles.Org").
		Where("id = ?", accountID).
		First(&account)

	if res.Error != nil {
		s.l.Error("failed to get account", zap.Error(res.Error), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch account: %v", res.Error)})
		return
	}

	// Get apps created by this account with config count
	type AppWithConfigCount struct {
		app.App
		ConfigCount int `gorm:"column:config_count"`
	}

	var appsWithCount []AppWithConfigCount
	appRes := s.db.WithContext(ctx).
		Model(&app.App{}).
		Select("apps.*, "+
			"(SELECT COUNT(*) FROM app_configs WHERE app_configs.app_id = apps.id) as config_count").
		Preload("Org").
		Where("apps.created_by_id = ?", accountID).
		Order("apps.created_at desc").
		Limit(100).
		Find(&appsWithCount)

	// Convert to []*app.App
	apps := make([]*app.App, len(appsWithCount))
	for i := range appsWithCount {
		appsWithCount[i].App.ConfigCount = appsWithCount[i].ConfigCount
		apps[i] = &appsWithCount[i].App
	}

	if appRes.Error != nil {
		s.l.Error("failed to get apps for account", zap.Error(appRes.Error), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch apps: %v", appRes.Error)})
		return
	}

	// Get installs created by this account with pagination
	installs, installsTotalPages, err := s.getInstallsForAccount(ctx, accountID, page)
	if err != nil {
		s.l.Error("failed to get installs for account", zap.Error(err), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch installs: %v", err)})
		return
	}

	// Parse entity type filters (comma-separated or multiple params)
	var entityTypes []string
	if typeFilter := c.Query("entity_types"); typeFilter != "" {
		// Split comma-separated values
		for _, t := range strings.Split(typeFilter, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				entityTypes = append(entityTypes, trimmed)
			}
		}
	}

	// Fetch audit logs
	auditLogs, auditLogsTotalPages, err := s.getAuditLogsForAccount(
		ctx, accountID, startDate, endDate, page, entityTypes,
	)
	if err != nil {
		s.l.Warn("failed to get audit logs", zap.Error(err))
		auditLogs = []*AuditLogEntry{}
		auditLogsTotalPages = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"account":                &account,
		"apps":                   apps,
		"installs":               installs,
		"audit_logs":             auditLogs,
		"start_date":             startDate,
		"end_date":               endDate,
		"page":                   page,
		"installs_total_pages":   installsTotalPages,
		"audit_logs_total_pages": auditLogsTotalPages,
	})
}

func (s *service) getInstallsForAccount(ctx context.Context, accountID string, page int) ([]*app.Install, int, error) {
	var installs []*app.Install
	var totalCount int64

	query := s.db.WithContext(ctx).
		Model(&app.Install{}).
		Unscoped().
		Where("created_by_id = ?", accountID)

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count installs: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(accountInstallsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate offset
	offset := (page - 1) * accountInstallsPerPage

	// Get paginated results
	res := query.
		Preload("Org").
		Preload("App").
		Preload("RunnerGroup.Runners").
		Preload("AppConfig").
		Preload("AppRunnerConfig").
		Order("created_at desc").
		Limit(accountInstallsPerPage).
		Offset(offset).
		Find(&installs)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, totalPages, nil
}

func (s *service) getAuditLogsForAccount(
	ctx context.Context,
	accountID string,
	startDate, endDate time.Time,
	page int,
	entityTypes []string,
) ([]*AuditLogEntry, int, error) {
	var entries []*AuditLogEntry

	// Build list of queries based on selected entity types (or all if none selected)
	allTypes := []string{"app", "workflow", "runner_job", "org", "app_sync"}
	if len(entityTypes) == 0 {
		entityTypes = allTypes
	}

	// Build individual queries for each entity type
	var queries []string
	var queryParams []interface{}

	for _, entityType := range entityTypes {
		switch entityType {
		case "app":
			queries = append(queries, `
				SELECT
					'app' as entity_type,
					a.id as entity_id,
					a.name as entity_name,
					a.created_at,
					a.org_id,
					o.name as org_name,
					NULL as app_id,
					NULL as app_name,
					a.description
				FROM apps a
				LEFT JOIN orgs o ON a.org_id = o.id
				WHERE a.created_by_id = ? AND a.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, accountID, startDate, endDate)

		case "workflow":
			queries = append(queries, `
				SELECT
					'workflow' as entity_type,
					aw.id as entity_id,
					aw.name as entity_name,
					aw.created_at,
					aw.org_id,
					o.name as org_name,
					aw.app_id,
					a.name as app_name,
					CAST(aw.status AS TEXT) as description
				FROM action_workflows aw
				LEFT JOIN orgs o ON aw.org_id = o.id
				LEFT JOIN apps a ON aw.app_id = a.id
				WHERE aw.created_by_id = ? AND aw.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, accountID, startDate, endDate)

		case "runner_job":
			queries = append(queries, `
				SELECT
					'runner_job' as entity_type,
					rj.id as entity_id,
					CONCAT(CAST(rj.type AS TEXT), ' - ', CAST(rj.operation AS TEXT)) as entity_name,
					rj.created_at,
					rj.org_id,
					o.name as org_name,
					NULL as app_id,
					NULL as app_name,
					rj.owner_type as description
				FROM runner_jobs rj
				LEFT JOIN orgs o ON rj.org_id = o.id
				WHERE rj.created_by_id = ? AND rj.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, accountID, startDate, endDate)

		case "org":
			queries = append(queries, `
				SELECT
					'org' as entity_type,
					o.id as entity_id,
					o.name as entity_name,
					o.created_at,
					NULL as org_id,
					NULL as org_name,
					NULL as app_id,
					NULL as app_name,
					NULL as description
				FROM orgs o
				WHERE o.created_by_id = ? AND o.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, accountID, startDate, endDate)

		case "app_sync":
			queries = append(queries, `
				SELECT
					'app_sync' as entity_type,
					ac.id as entity_id,
					'App Sync' as entity_name,
					ac.created_at,
					ac.org_id,
					o.name as org_name,
					ac.app_id,
					a.name as app_name,
					ac.cli_version as description
				FROM app_configs ac
				LEFT JOIN orgs o ON ac.org_id = o.id
				LEFT JOIN apps a ON ac.app_id = a.id
				WHERE ac.created_by_id = ? AND ac.created_at BETWEEN ? AND ?
			`)
			queryParams = append(queryParams, accountID, startDate, endDate)
		}
	}

	if len(queries) == 0 {
		return []*AuditLogEntry{}, 1, nil
	}

	// Join all queries with UNION ALL
	query := ""
	for i, q := range queries {
		if i > 0 {
			query += " UNION ALL "
		}
		query += q
	}
	query += " ORDER BY created_at DESC"

	// Get total count with separate query
	countQuery := `SELECT COUNT(*) FROM (` + query + `) as audit_entries`

	var totalCount int64
	err := s.db.WithContext(ctx).Raw(countQuery, queryParams...).Scan(&totalCount).Error

	if err != nil {
		return nil, 0, fmt.Errorf("unable to count audit logs: %w", err)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(totalCount) / float64(accountAuditLogsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * accountAuditLogsPerPage

	// Execute paginated query
	queryParams = append(queryParams, accountAuditLogsPerPage, offset)
	err = s.db.WithContext(ctx).Raw(query+` LIMIT ? OFFSET ?`, queryParams...).Scan(&entries).Error

	if err != nil {
		return nil, 0, fmt.Errorf("unable to get audit logs: %w", err)
	}

	return entries, totalPages, nil
}
