package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallConfigSyncs
// @Summary				get config sync history for an install
// @Description			Returns the install config sync history, ordered by most recent first.
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.InstallConfigSync
// @Router					/v1/installs/{install_id}/config-syncs [GET]
func (s *service) GetInstallConfigSyncs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	var install app.Install
	if err := s.db.WithContext(ctx).First(&install, "id = ? AND org_id = ?", installID, org.ID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	var syncs []app.InstallConfigSync
	res := s.db.WithContext(ctx).
		Preload("Versions", "install_id = ?", installID).
		Preload("VCSConnectionCommit").
		Where(app.InstallConfigSync{
			OrgID: org.ID,
		}).
		Joins("JOIN install_config_versions ON install_config_versions.install_config_sync_id = install_config_syncs.id AND install_config_versions.install_id = ?", installID).
		Group("install_config_syncs.id").
		Order("install_config_syncs.created_at DESC").
		Limit(50).
		Find(&syncs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install config syncs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, syncs)
}
