package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const orgsPerPage = 8

func (s *service) Orgs(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	tagFilters := c.QueryArray("tag")

	// Split comma-separated values and filter out empty strings
	var filteredTags []string
	for _, tag := range tagFilters {
		if tag != "" {
			// Split on comma in case multiple tags come as a single value
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					filteredTags = append(filteredTags, trimmed)
				}
			}
		}
	}

	page := getPageFromQuery(c)

	orgs, totalPages, err := s.getOrgs(ctx, search, filteredTags, page)
	if err != nil {
		s.l.Error("failed to get orgs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	allTags, err := s.getAllOrgTags(ctx)
	if err != nil {
		s.l.Warn("failed to get all tags", zap.Error(err))
		allTags = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"orgs":        orgs,
		"all_tags":    allTags,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) getOrgs(ctx context.Context, search string, tagFilters []string, page int) ([]*app.Org, int, error) {
	type OrgWithCounts struct {
		app.Org
		AppCount     int `gorm:"column:app_count"`
		InstallCount int `gorm:"column:install_count"`
	}

	var orgsWithCounts []OrgWithCounts
	var totalCount int64

	// Build base query
	query := s.db.WithContext(ctx).Model(&app.Org{})

	// Apply search filter if provided
	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"name ILIKE ? OR id = ?",
			"%"+search+"%",
			search,
		)
	}

	// Apply tag filters if provided (org must have ANY of the selected tags - OR logic)
	if len(tagFilters) > 0 {
		// Use && operator with explicit text[] cast for array overlap (OR logic)
		query = query.Where("tags && CAST(? AS text[])", pq.Array(tagFilters))
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count orgs: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(orgsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate offset
	offset := (page - 1) * orgsPerPage

	// Get paginated results with counts (excluding deleted records)
	res := query.
		Select("orgs.*, " +
			"(SELECT COUNT(*) FROM apps WHERE apps.org_id = orgs.id AND apps.deleted_at = 0) as app_count, " +
			"(SELECT COUNT(*) FROM installs WHERE installs.org_id = orgs.id AND installs.deleted_at = 0) as install_count").
		Order("created_at desc").
		Limit(orgsPerPage).
		Offset(offset).
		Find(&orgsWithCounts)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	// Convert to []*app.Org with counts embedded
	orgs := make([]*app.Org, len(orgsWithCounts))
	for i := range orgsWithCounts {
		orgsWithCounts[i].Org.AppCount = orgsWithCounts[i].AppCount
		orgsWithCounts[i].Org.InstallCount = orgsWithCounts[i].InstallCount
		orgs[i] = &orgsWithCounts[i].Org
	}

	return orgs, totalPages, nil
}

func (s *service) getAllOrgTags(ctx context.Context) ([]string, error) {
	var tags []string
	err := s.db.WithContext(ctx).
		Model(&app.Org{}).
		Distinct().
		Pluck("unnest(tags)", &tags).
		Error
	if err != nil {
		return nil, fmt.Errorf("unable to get tags: %w", err)
	}
	return tags, nil
}
