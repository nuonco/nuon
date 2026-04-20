package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetLatestInstallRoles
// @Summary				get latest install roles
// @Description			get install roles for the current app config
// @Tags					installs
// @Param					install_id	path	string	true	"install ID"
// @Param					offset		query	int		false	"offset of results to return"	Default(0)
// @Param					limit		query	int		false	"limit of results to return"	Default(10)
// @Param					page		query	int		false	"page number of results to return"	Default(0)
// @Param					q			query	string	false	"search query to filter roles by display name"
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
// @Router					/v1/installs/{install_id}/roles/latest [get]
func (s *service) GetLatestInstallRoles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	var install app.Install
	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", installID, org.ID).
		First(&install)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", res.Error))
		return
	}

	q := ctx.Query("q")

	var roles []app.InstallRoles
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("AppRoleConfig").
		Joins("JOIN app_awsiam_role_configs ON app_awsiam_role_configs.id = install_roles.app_role_config_id AND app_awsiam_role_configs.app_config_id = ?", install.AppConfigID).
		Where("install_roles.install_id = ? AND install_roles.org_id = ?", installID, org.ID).
		Order("install_roles.created_at DESC")

	if q != "" {
		tx = tx.Where("app_awsiam_role_configs.display_name ILIKE ?", "%"+q+"%")
	}

	res = tx.Find(&roles)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install roles: %w", res.Error))
		return
	}

	roles, err = db.HandlePaginatedResponse(ctx, roles)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, roles)
}
