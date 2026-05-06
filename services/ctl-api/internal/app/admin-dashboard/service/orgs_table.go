package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrgsTable returns just the orgs table data as JSON
func (s *service) OrgsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	label := c.Query("label")
	page := getPageFromQuery(c)

	orgs, totalPages, err := s.getOrgs(ctx, search, label, page)
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
