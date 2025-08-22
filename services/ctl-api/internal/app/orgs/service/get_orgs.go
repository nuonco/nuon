package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID                     GetOrgs
// @Summary                Return current user's orgs
// @Description.markdown   get_orgs.md
// @Param                  q                           query   string  false   "search query"
// @Param                  offset                      query   int     false   "offset of results to return"   Default(0)
// @Param                  limit                       query   int     false   "limit of results to return"    Default(10)
// @Param                  page                        query   int     false   "page number of results to return"   Default(0)
// @Param                  x-nuon-pagination-enabled   header  bool    false   "Enable pagination"
// @Tags                   orgs
// @Accept                 json
// @Produce                json
// @Security               APIKey
// @Failure                400 {object} stderr.ErrResponse
// @Failure                401 {object} stderr.ErrResponse
// @Failure                403 {object} stderr.ErrResponse
// @Failure                404 {object} stderr.ErrResponse
// @Failure                500 {object} stderr.ErrResponse
// @Success                200 {array}  app.Org
// @Router                 /v1/orgs [GET]
func (s *service) GetCurrentUserOrgs(ctx *gin.Context) {
	q := ctx.Query("q")
	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	orgs, err := s.getOrgs(ctx, account.OrgIDs, q)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getOrgs(ctx *gin.Context, orgIDs []string, q string) ([]app.Org, error) {
	var orgs []app.Org
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Joins("JOIN accounts ON accounts.id = orgs.created_by_id").
		Where("orgs.id IN ?", orgIDs).
		Order(fmt.Sprintf("CASE WHEN accounts.account_type = '%s' THEN 1 ELSE 0 END, orgs.id", app.AccountTypeCanary))

	if q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}

	res := tx.
		Find(&orgs)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get orgs")
	}

	orgs, err := db.HandlePaginatedResponse(ctx, orgs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to handle paginated response")
	}

	return orgs, nil
}
