package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
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
	role := app.Role{}
	res := s.db.WithContext(ctx).
		Where("org_id = ? AND role_type = ?", orgID, app.RoleTypeOrgAdmin).
		First(&role)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org accounts %s: %w", orgID, res.Error)
	}

	ar := []app.AccountRole{}
	tx := s.db.WithContext(ctx)

	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get account from context: %w", err)
	}

	if !strings.HasSuffix(acct.Email, "nuon.co") {
		tx = tx.Joins("JOIN accounts ON accounts.id = account_roles.account_id")
		tx = tx.Where("accounts.email NOT LIKE ?", "%nuon.co")
	}

	tx = tx.
		Scopes(scopes.WithOffsetPagination).
		Preload("Account").
		Where("role_id = ?", role.ID).
		Find(&ar)

	ar, err = db.HandlePaginatedResponse(ctx, ar)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	accounts := make([]app.Account, len(ar))
	for i, a := range ar {
		accounts[i] = a.Account
	}

	return accounts, nil
}
