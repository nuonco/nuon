package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InstallsTableGlobal returns the global installs table data as JSON
func (s *service) InstallsTableGlobal(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	sort := c.Query("sort")
	filter := c.Query("filter")
	deletedFilter := c.Query("deleted_filter")
	if deletedFilter == "" {
		deletedFilter = "active"
	}
	page := getPageFromQuery(c)

	installs, totalPages, err := s.getInstalls(ctx, search, sort, filter, deletedFilter, page)
	if err != nil {
		s.l.Error("failed to get installs for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"installs":    installs,
		"page":        page,
		"total_pages": totalPages,
	})
}
