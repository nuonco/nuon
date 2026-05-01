package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccountsTable returns just the accounts table data as JSON
func (s *service) AccountsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	filter := c.Query("filter")
	page := getPageFromQuery(c)

	accounts, totalPages, err := s.getAccounts(ctx, search, filter, page)
	if err != nil {
		s.l.Error("failed to get accounts for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":    accounts,
		"page":        page,
		"total_pages": totalPages,
	})
}
