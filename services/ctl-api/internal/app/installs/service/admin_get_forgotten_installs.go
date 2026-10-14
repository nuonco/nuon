package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						AdminGetForgottenInstalls
// @Summary				list forgotten installs for an org
// @Description.markdown	admin_get_forgotten_installs.md
// @Param					org_id	path	string	true	"org ID"
// @Param					offset	query	int		false	"offset of results to return"		Default(0)
// @Param					limit	query	int		false	"limit of results to return"		Default(10)
// @Param					page	query	int		false	"page number of results to return"	Default(0)
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Install
// @Router					/v1/orgs/{org_id}/admin-get-forgotten-installs [GET]
func (s *service) AdminGetForgottenInstalls(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	installs, err := s.getForgottenOrgInstalls(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get forgotten installs for org %s: %w", orgID, err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getForgottenOrgInstalls(ctx *gin.Context, orgID string) ([]app.Install, error) {
	var installs []app.Install
	res := s.db.WithContext(ctx).
		Unscoped().
		Scopes(scopes.WithDisableViews, scopes.WithOffsetPagination).
		Where(app.Install{OrgID: orgID}).
		Where("deleted_at > ?", 0).
		Order("deleted_at DESC, id DESC").
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get forgotten installs: %w", res.Error)
	}

	installs, err := db.HandlePaginatedResponse(ctx, installs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}
	return installs, nil
}
