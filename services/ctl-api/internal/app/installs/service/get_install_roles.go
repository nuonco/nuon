package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallRoles
// @Summary				get install roles
// @Description			get all roles for an install
// @Tags					installs
// @Param					install_id	path	string	true	"install ID"
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallRoles
// @Router					/v1/installs/{install_id}/roles [get]
func (s *service) GetInstallRoles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	var roles []app.InstallRoles
	res := s.db.WithContext(ctx).
		Preload("AppRoleConfig").
		Where("install_id = ? AND org_id = ?", installID, org.ID).
		Find(&roles)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install roles: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, roles)
}
