package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) AccountInstallsTable(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("id")
	page := getPageFromQuery(c)

	installs, installsTotalPages, err := s.getInstallsForAccount(ctx, accountID, page)
	if err != nil {
		s.l.Error("failed to get installs for account table", zap.Error(err), zap.String("account_id", accountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch installs: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id":  accountID,
		"installs":    installs,
		"page":        page,
		"total_pages": installsTotalPages,
	})
}
