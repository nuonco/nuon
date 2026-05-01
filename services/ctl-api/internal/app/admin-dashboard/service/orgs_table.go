package service

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrgsTable returns just the orgs table data as JSON
func (s *service) OrgsTable(c *gin.Context) {
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
		s.l.Error("failed to get orgs for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orgs":        orgs,
		"page":        page,
		"total_pages": totalPages,
	})
}
