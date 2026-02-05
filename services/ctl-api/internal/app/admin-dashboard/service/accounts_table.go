package service

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

// AccountsTable returns just the accounts table for htmx polling
func (s *service) AccountsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")

	accounts, err := s.getAccounts(ctx, search)
	if err != nil {
		s.l.Error("failed to get accounts for table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	component := views.AccountsTable(accounts)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
