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
	feature := c.Query("feature")
	featureState := c.Query("feature_state")
	page := getPageFromQuery(c)

	orgs, featureCounts, totalPages, err := s.getOrgs(ctx, search, label, feature, featureState, page)
	if err != nil {
		s.l.Error("failed to get orgs for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orgs":           orgs,
		"feature_counts": featureCounts,
		"page":           page,
		"total_pages":    totalPages,
	})
}
