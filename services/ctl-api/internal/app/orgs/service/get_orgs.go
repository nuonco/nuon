package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID GetOrgs
// @Summary	Return current user's orgs
// @Description.markdown get_orgs.md
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.Org
// @Router			/v1/orgs [GET]
func (s *service) GetCurrentUserOrgs(ctx *gin.Context) {
	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	orgs, err := s.getOrgs(ctx, account.OrgIDs)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getOrgs(ctx context.Context, orgIDs []string) ([]app.Org, error) {
	var orgs []app.Org
	res := s.db.WithContext(ctx).
		Joins("JOIN accounts ON accounts.id = orgs.created_by_id").
		Where("orgs.id IN ?", orgIDs).
		Order(fmt.Sprintf("CASE WHEN accounts.account_type = '%s' THEN 1 ELSE 0 END, orgs.id", app.AccountTypeCanary)).
		Find(&orgs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get orgs")
	}

	return orgs, nil
}
