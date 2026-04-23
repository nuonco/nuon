package service

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const labelsPerPage = 20

// LabelsPage renders the full labels browse page.
func (s *service) LabelsPage(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	entityType := c.Query("entity_type")
	orgID := c.Query("org_id")
	page := getPageFromQuery(c)

	results, allKeys, totalPages, err := s.getLabelsData(ctx, search, entityType, orgID, page)
	if err != nil {
		s.l.Error("failed to get labels data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels data"})
		return
	}

	orgs := s.getOrgOptions(ctx)

	component := views.LabelsPage(results, allKeys, orgs, search, entityType, orgID, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// LabelsTable renders just the labels table fragment for HTMX swaps.
func (s *service) LabelsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	entityType := c.Query("entity_type")
	orgID := c.Query("org_id")
	page := getPageFromQuery(c)

	results, _, totalPages, err := s.getLabelsData(ctx, search, entityType, orgID, page)
	if err != nil {
		s.l.Error("failed to get labels data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels data"})
		return
	}

	component := views.LabelsTableFragment(results, search, entityType, orgID, page, totalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

// LabelSearchResult represents a single entity with labels for the browse page.
type LabelSearchResult = views.LabelSearchResult

func (s *service) getLabelsData(ctx context.Context, search, entityType, orgID string, page int) ([]views.LabelSearchResult, []string, int, error) {
	// 1. Get all distinct label keys across all entity types
	allKeys := s.getAllLabelKeys(ctx)

	// 2. Query entities matching the search/filter
	var results []views.LabelSearchResult

	type queryFunc func(context.Context, string, string) ([]views.LabelSearchResult, error)
	tables := []struct {
		entityType string
		query      queryFunc
	}{
		{"install", s.getInstallLabelResults},
		{"component", s.getComponentLabelResults},
		{"action", s.getActionLabelResults},
	}

	for _, t := range tables {
		if entityType != "" && entityType != t.entityType {
			continue
		}
		rows, err := t.query(ctx, search, orgID)
		if err != nil {
			s.l.Warn("failed to get label results for "+t.entityType, zap.Error(err))
			continue
		}
		results = append(results, rows...)
	}

	// 3. Paginate
	totalCount := len(results)
	totalPages := int(math.Ceil(float64(totalCount) / float64(labelsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * labelsPerPage
	if offset > len(results) {
		offset = len(results)
	}
	end := offset + labelsPerPage
	if end > len(results) {
		end = len(results)
	}
	results = results[offset:end]

	return results, allKeys, totalPages, nil
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

func (s *service) getInstallLabelResults(ctx context.Context, search, orgID string) ([]views.LabelSearchResult, error) {
	var installs []app.Install
	tx := s.db.WithContext(ctx).
		Where("labels IS NOT NULL AND labels != '{}'::jsonb")

	if orgID != "" {
		tx = tx.Where("org_id = ?", orgID)
	}
	tx = applyLabelSearch(tx, search)

	if err := tx.Find(&installs).Error; err != nil {
		return nil, err
	}

	results := make([]views.LabelSearchResult, 0, len(installs))
	for _, i := range installs {
		results = append(results, views.LabelSearchResult{
			EntityType: "install",
			EntityID:   i.ID,
			EntityName: i.Name,
			Labels:     i.Labels,
			DetailURL:  "/installs/" + i.ID,
		})
	}
	return results, nil
}

func (s *service) getComponentLabelResults(ctx context.Context, search, orgID string) ([]views.LabelSearchResult, error) {
	var components []app.Component
	tx := s.db.WithContext(ctx).
		Where("labels IS NOT NULL AND labels != '{}'::jsonb")

	if orgID != "" {
		tx = tx.Where("app_id IN (SELECT id FROM apps WHERE org_id = ?)", orgID)
	}
	tx = applyLabelSearch(tx, search)

	if err := tx.Find(&components).Error; err != nil {
		return nil, err
	}

	results := make([]views.LabelSearchResult, 0, len(components))
	for _, c := range components {
		results = append(results, views.LabelSearchResult{
			EntityType: "component",
			EntityID:   c.ID,
			EntityName: c.Name,
			Labels:     c.Labels,
		})
	}
	return results, nil
}

func (s *service) getActionLabelResults(ctx context.Context, search, orgID string) ([]views.LabelSearchResult, error) {
	var actions []app.ActionWorkflow
	tx := s.db.WithContext(ctx).
		Where("labels IS NOT NULL AND labels != '{}'::jsonb")

	if orgID != "" {
		tx = tx.Where("org_id = ?", orgID)
	}
	tx = applyLabelSearch(tx, search)

	if err := tx.Find(&actions).Error; err != nil {
		return nil, err
	}

	results := make([]views.LabelSearchResult, 0, len(actions))
	for _, a := range actions {
		results = append(results, views.LabelSearchResult{
			EntityType: "action",
			EntityID:   a.ID,
			EntityName: a.Name,
			Labels:     a.Labels,
		})
	}
	return results, nil
}

// applyLabelSearch adds WHERE clauses for label search.
// Supports "key:value" (exact match) or just "key" (key existence).
func applyLabelSearch(tx *gorm.DB, search string) *gorm.DB {
	search = strings.TrimSpace(search)
	if search == "" {
		return tx
	}

	lbls := labels.ParseLabelsQuery(search)
	if lbls != nil {
		return tx.Scopes(labels.WithLabels("labels", lbls))
	}

	// Treat as key-only search (check if key exists in JSONB)
	return tx.Where("jsonb_exists(labels, ?)", search)
}
