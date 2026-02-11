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
	var orgs []*app.Org
	var totalCount int64

	// Build base query
	query := s.db.WithContext(ctx).Model(&app.Org{})

	// Apply search filter if provided
	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"name ILIKE ? OR id = ? OR ? = ANY(tags)",
			"%"+search+"%",
			search,
			search,
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

	// Get paginated results
	res := query.
		Order("created_at desc").
		Limit(orgsPerPage).
		Offset(offset).
		Find(&orgs)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	return orgs, totalPages, nil
}
