package service

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const accountsPerPage = 8

func (s *service) Accounts(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	filter := c.Query("filter")
	page := getPageFromQuery(c)

	accounts, totalPages, err := s.getAccounts(ctx, search, filter, page)
	if err != nil {
		s.l.Error("failed to get accounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":    accounts,
		"page":        page,
		"total_pages": totalPages,
	})
}

func (s *service) getAccounts(ctx context.Context, search string, filter string, page int) ([]*app.Account, int, error) {
	var accounts []*app.Account
	var totalCount int64

	query := s.readDB().WithContext(ctx).
		Model(&app.Account{}).
		Preload("Roles.Org").
		Where("account_type IN ?", []app.AccountType{app.AccountTypeAuth0, app.AccountTypeAuth})

	switch filter {
	case "nuon":
		query = query.Where("email LIKE ?", "%@nuon.co")
	case "user":
		query = query.Where("email NOT LIKE ?", "%@nuon.co")
	}

	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where(
			"email ILIKE ? OR id = ?",
			"%"+search+"%",
			search,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count accounts: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(accountsPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * accountsPerPage

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
