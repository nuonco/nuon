package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const installsPerPage = 8

func (s *service) Installs(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	page := getPageFromQuery(c)

	installs, totalPages, err := s.getInstalls(ctx, search, page)
	if err != nil {
		s.l.Error("failed to get installs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	component := views.Installs(installs, page, totalPages, search)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getInstalls(ctx context.Context, search string, page int) ([]*app.Install, int, error) {
	var installs []*app.Install
	var totalCount int64

	// Build base query
	query := s.db.WithContext(ctx).Model(&app.Install{})

	// Apply search filter if provided
	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"name ILIKE ? OR id = ?",
			"%"+search+"%",
			search,
		)
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count installs: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(installsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate offset
	offset := (page - 1) * installsPerPage

	// Get paginated results
	res := query.
		Preload("Org").
		Preload("App").
		Preload("RunnerGroup").
		Order("created_at desc").
		Limit(installsPerPage).
		Offset(offset).
		Find(&installs)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get installs: %w", res.Error)
	}

	return installs, totalPages, nil
}

func getPageFromQuery(c *gin.Context) int {
	pageStr := c.Query("page")
	if pageStr == "" {
		return 1
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1
	}
	return page
}
