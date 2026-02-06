package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) Accounts(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")

	accounts, err := s.getAccounts(ctx, search)
	if err != nil {
		s.l.Error("failed to get accounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	component := views.Accounts(accounts)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getAccounts(ctx context.Context, search string) ([]*app.Account, error) {
	var accounts []*app.Account

	query := s.db.WithContext(ctx).
		Preload("Roles.Org").
		Where("account_type IN ?", []app.AccountType{app.AccountTypeAuth0, app.AccountTypeAuth})

	// Apply search filter if provided
	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"email ILIKE ? OR id = ?",
			"%"+search+"%",
			search,
		)
	}

	res := query.
		Order("created_at desc").
		Limit(100).
		Find(&accounts)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get accounts: %w", res.Error)
	}

	return accounts, nil
}
