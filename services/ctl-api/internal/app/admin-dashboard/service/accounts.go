package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const accountsPerPage = 8

func (s *service) Accounts(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	page := getPageFromQuery(c)

	accounts, totalPages, err := s.getAccounts(ctx, search, page)
	if err != nil {
		s.l.Error("failed to get accounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	component := views.Accounts(accounts, page, totalPages, search)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getAccounts(ctx context.Context, search string, page int) ([]*app.Account, int, error) {
	var accounts []*app.Account
	var totalCount int64

	// Build base query
	query := s.db.WithContext(ctx).
		Model(&app.Account{}).
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

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count accounts: %w", err)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(accountsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Calculate offset
	offset := (page - 1) * accountsPerPage

	// Get paginated results
	res := query.
		Order("created_at desc").
		Limit(accountsPerPage).
		Offset(offset).
		Find(&accounts)

	if res.Error != nil {
		return nil, 0, fmt.Errorf("unable to get accounts: %w", res.Error)
	}

	return accounts, totalPages, nil
}
