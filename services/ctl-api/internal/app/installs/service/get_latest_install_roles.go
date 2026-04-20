package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetLatestInstallRoles
// @Summary				get latest install roles
// @Description			get install roles for the current app config
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

	var roles []app.InstallRoles
	res = s.db.WithContext(ctx).
		Preload("AppRoleConfig", func(db *gorm.DB) *gorm.DB {
			return db.Where("app_config_id = ?", install.AppConfigID)
		}).
		Where("install_id = ? AND org_id = ?", installID, org.ID).
		Find(&roles)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install roles: %w", res.Error))
		return
	}

	// filter to only roles that matched the current app config
	latest := make([]app.InstallRoles, 0, len(roles))
	for _, r := range roles {
		if r.AppRoleConfig.ID != "" {
			latest = append(latest, r)
		}
	}

	ctx.JSON(http.StatusOK, latest)
}
