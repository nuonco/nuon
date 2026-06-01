package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgAcounts
// @Summary				Get user accounts for current org
// @Description.markdown	get_org.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/orgs/current/accounts [GET]
func (s *service) GetOrgAccounts(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accounts, err := s.getOrgAccounts(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

func (s *service) getOrgAccounts(ctx *gin.Context, orgID string) ([]app.Account, error) {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get account from context: %w", err)
	}

	accounts := []app.Account{}

	// Drive from the org's membership (indexed account_roles.org_id) rather than
	// scanning the accounts table via an IN-subquery + LIMIT.
	tx := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Joins("JOIN account_roles ON account_roles.account_id = accounts.id AND account_roles.org_id = ? AND account_roles.deleted_at = 0", orgID).
		Where("accounts.account_type != ?", app.AccountTypeService).
		Group("accounts.id").
		Order("accounts.email").
		Order("accounts.id")

	if !strings.HasSuffix(acct.Email, "nuon.co") {
		tx = tx.Where("accounts.email NOT LIKE ?", "%nuon.co")
	}

	tx = tx.
		Scopes(scopes.WithOffsetPagination).
		Preload("Roles", "org_id = ?", orgID).
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Find(&accounts)
	if tx.Error != nil {
		return nil, fmt.Errorf("unable to get org accounts %s: %w", orgID, tx.Error)
	}

	accounts, err = db.HandlePaginatedResponse(ctx, accounts)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return accounts, nil
}
