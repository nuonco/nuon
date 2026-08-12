package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetRole
// @Summary				Get one of your org's roles
// @Description			Get a role, including the scoped permission entries its policy carries.
// @Param					role_id	path	string	true	"role ID"
// @Tags					roles
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Role
// @Router					/v1/roles/{role_id} [GET]
func (s *service) GetRole(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	role, err := s.getOrgRole(ctx, org.ID, ctx.Param("role_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, role)
}
