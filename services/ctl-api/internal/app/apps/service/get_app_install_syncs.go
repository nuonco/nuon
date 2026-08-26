package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppInstallSyncs
// @Summary				list app install config syncs
// @Description			Returns a list of app install config sync records for the given app.
// @Tags					apps
// @Produce				json
// @Security				APIKey && OrgID
// @Param					app_id	path	string	true	"app ID"
// @Success				200	{array}		app.AppInstallConfigSync
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/install-syncs [get]
func (s *service) GetAppInstallSyncs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")

	var syncs []app.AppInstallConfigSync
	if err := s.db.WithContext(ctx).
		Where(app.AppInstallConfigSync{AppID: appID, OrgID: org.ID}).
		Preload("InstallConfigSyncs").
		Preload("VCSConnectionCommit").
		Preload("InstallCreationApproval").
		Order("created_at DESC").
		Limit(50).
		Find(&syncs).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install syncs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, syncs)
}
