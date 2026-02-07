package service

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
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

	component := views.AccountInstallsTable(accountID, installs, page, installsTotalPages)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
