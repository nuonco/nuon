package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const orgsPerPage = 8

func (s *service) Orgs(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	page := getPageFromQuery(c)

	orgs, totalPages, err := s.getOrgs(ctx, search, page)
	if err != nil {
		s.l.Error("failed to get orgs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	component := views.Orgs(orgs, page, totalPages, search)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getOrgs(ctx context.Context, search string, page int) ([]*app.Org, int, error) {
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
			"name ILIKE ? OR id = ? OR EXISTS (SELECT 1 FROM unnest(tags) AS tag WHERE tag ILIKE ?)",
			"%"+search+"%",
			search,
			"%"+search+"%",
		)
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

	// Get paginated results with counts
	res := query.
		Select("orgs.*, " +
			"(SELECT COUNT(*) FROM apps WHERE apps.org_id = orgs.id) as app_count, " +
			"(SELECT COUNT(*) FROM installs WHERE installs.org_id = orgs.id) as install_count").
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
