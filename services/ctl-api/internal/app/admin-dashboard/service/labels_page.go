package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const labelsPerPage = 20

// LabelsPage returns the labels browse data.
func (s *service) LabelsPage(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	entityType := c.Query("entity_type")
	orgID := c.Query("org_id")
	page := getPageFromQuery(c)

	results, allKeys, totalPages, totalCount, err := s.getLabelsData(ctx, search, entityType, orgID, page)
	if err != nil {
		s.l.Error("failed to get labels data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels data"})
		return
	}

	if results == nil {
		results = []views.LabelSearchResult{}
	}
	if allKeys == nil {
		allKeys = []string{}
	}

	orgs := s.getOrgOptions(ctx)
	if orgs == nil {
		orgs = []views.OrgOption{}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"all_keys":    allKeys,
		"orgs":        orgs,
		"page":        page,
		"total_pages": totalPages,
		"total_count": totalCount,
	})
}

// LabelsTable returns just the labels table data.
func (s *service) LabelsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	entityType := c.Query("entity_type")
	orgID := c.Query("org_id")
	page := getPageFromQuery(c)

	results, _, totalPages, totalCount, err := s.getLabelsData(ctx, search, entityType, orgID, page)
	if err != nil {
		s.l.Error("failed to get labels data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     results,
		"page":        page,
		"total_pages": totalPages,
		"total_count": totalCount,
	})
}

// LabelSearchResult represents a single entity with labels for the browse page.
type LabelSearchResult = views.LabelSearchResult

// labelTable describes one of the labelled tables that the labels-browse query
// unions over. Per-table org filtering varies because components are scoped via
// their parent app, not directly by org.
type labelTable struct {
	table         string
	entityType    string
	detailURLExpr string
	orgClause     string
}

var allLabelTables = []labelTable{
	{
		table:         "installs",
		entityType:    "install",
		detailURLExpr: "'/installs/' || id",
		orgClause:     "org_id = ?",
	},
	{
		table:         "components",
		entityType:    "component",
		detailURLExpr: "''",
		orgClause:     "app_id IN (SELECT id FROM apps WHERE org_id = ? AND deleted_at = 0)",
	},
	{
		table:         "action_workflows",
		entityType:    "action",
		detailURLExpr: "''",
		orgClause:     "org_id = ?",
	},
}

func labelTablesFor(entityType string) []labelTable {
	if entityType == "" {
		return allLabelTables
	}
	for _, t := range allLabelTables {
		if t.entityType == entityType {
			return []labelTable{t}
		}
	}
	return nil
}

// buildLabelFilterClauses turns a search string ("key:value", "key=value",
// "k1:v1,k2:*", or a bare key) into SQL clause fragments and args that match
// the same semantics as labels.WithLabels.
func buildLabelFilterClauses(search string) ([]string, []any) {
	search = strings.TrimSpace(search)
	if search == "" {
		return nil, nil
	}

	lbls := labels.ParseLabelsQuery(search)
	if lbls == nil {
		return []string{"jsonb_exists(labels, ?)"}, []any{search}
	}

	var clauses []string
	var args []any

	exact := make(labels.Labels)
	var wildcardKeys []string
	for k, v := range lbls {
		if v == "*" {
			wildcardKeys = append(wildcardKeys, k)
		} else {
			exact[k] = v
		}
	}

	if len(exact) > 0 {
		jsonBytes, err := json.Marshal(exact)
		if err == nil {
			clauses = append(clauses, "labels @> ?::jsonb")
			args = append(args, string(jsonBytes))
		}
	}

	sort.Strings(wildcardKeys)
	for _, key := range wildcardKeys {
		clauses = append(clauses, "jsonb_exists(labels, ?)")
		args = append(args, key)
	}

	return clauses, args
}

func (s *service) getLabelsData(ctx context.Context, search, entityType, orgID string, page int) ([]views.LabelSearchResult, []string, int, int64, error) {
	allKeys := s.getAllLabelKeys(ctx)

	tables := labelTablesFor(entityType)
	if len(tables) == 0 {
		return []views.LabelSearchResult{}, allKeys, 1, 0, nil
	}

	labelClauses, labelArgs := buildLabelFilterClauses(search)

	var dataSubqueries []string
	var countSubqueries []string
	var dataArgs []any
	var countArgs []any

	for _, t := range tables {
		clauses := []string{
			"deleted_at = 0",
			"labels IS NOT NULL",
			"labels != '{}'::jsonb",
		}
		var args []any
		if orgID != "" {
			clauses = append(clauses, t.orgClause)
			args = append(args, orgID)
		}
		clauses = append(clauses, labelClauses...)
		args = append(args, labelArgs...)

		where := strings.Join(clauses, " AND ")

		dataSubqueries = append(dataSubqueries, fmt.Sprintf(
			"SELECT '%s' AS entity_type, id AS entity_id, name AS entity_name, labels, created_at, %s AS detail_url FROM %s WHERE %s",
			t.entityType, t.detailURLExpr, t.table, where,
		))
		countSubqueries = append(countSubqueries, fmt.Sprintf(
			"SELECT 1 FROM %s WHERE %s", t.table, where,
		))

		dataArgs = append(dataArgs, args...)
		countArgs = append(countArgs, args...)
	}

	countSQL := "SELECT COUNT(*) FROM (" + strings.Join(countSubqueries, " UNION ALL ") + ") c"
	var totalCount int64
	if err := s.db.WithContext(ctx).Raw(countSQL, countArgs...).Scan(&totalCount).Error; err != nil {
		return nil, allKeys, 1, 0, fmt.Errorf("unable to count labels: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(labelsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * labelsPerPage
	dataSQL := "SELECT entity_type, entity_id, entity_name, labels, detail_url FROM (" +
		strings.Join(dataSubqueries, " UNION ALL ") +
		") u ORDER BY u.created_at DESC LIMIT ? OFFSET ?"
	dataArgs = append(dataArgs, labelsPerPage, offset)

	var results []views.LabelSearchResult
	if err := s.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
		return nil, allKeys, 1, 0, fmt.Errorf("unable to query labels: %w", err)
	}

	if results == nil {
		results = []views.LabelSearchResult{}
	}

	return results, allKeys, totalPages, totalCount, nil
}

func (s *service) getAllLabelKeys(ctx context.Context) []string {
	var keys []string

	// Query each table separately so a failure in one doesn't break the whole page.
	tables := []string{"installs", "components", "action_workflows"}
	for _, table := range tables {
		var tableKeys []string
		query := "SELECT DISTINCT (jsonb_each_text(labels)).key FROM " + table +
			" WHERE labels IS NOT NULL AND labels != '{}'::jsonb AND deleted_at = 0"
		if err := s.db.WithContext(ctx).Raw(query).Scan(&tableKeys).Error; err != nil {
			s.l.Warn("failed to get label keys from "+table, zap.Error(err))
			continue
		}
		keys = append(keys, tableKeys...)
	}

	// Deduplicate and sort.
	seen := make(map[string]bool)
	unique := make([]string, 0, len(keys))
	for _, k := range keys {
		if !seen[k] {
			seen[k] = true
			unique = append(unique, k)
		}
	}
	sort.Strings(unique)
	return unique
}

func (s *service) getOrgOptions(ctx context.Context) []views.OrgOption {
	var orgs []app.Org
	if err := s.db.WithContext(ctx).Select("id", "name").Order("name ASC").Find(&orgs).Error; err != nil {
		s.l.Warn("failed to get orgs for labels filter", zap.Error(err))
		return nil
	}

	opts := make([]views.OrgOption, 0, len(orgs))
	for _, o := range orgs {
		opts = append(opts, views.OrgOption{ID: o.ID, Name: o.Name})
	}
	return opts
}
