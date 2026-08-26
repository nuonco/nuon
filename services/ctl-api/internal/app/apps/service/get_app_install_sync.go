package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppInstallSync
// @Summary				get a single app install config sync
// @Description			Returns a single app install config sync record with child install config syncs.
// @Tags					apps
// @Produce				json
// @Security				APIKey && OrgID
// @Param					app_id		path	string	true	"app ID"
// @Param					sync_id		path	string	true	"sync ID"
// @Success				200	{object}	app.AppInstallConfigSync
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/apps/{app_id}/install-syncs/{sync_id} [get]
func (s *service) GetAppInstallSync(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.requireInstallSyncing(ctx) {
		return
	}

	appID := ctx.Param("app_id")
	syncID := ctx.Param("sync_id")

	var sync app.AppInstallConfigSync
	if err := s.db.WithContext(ctx).
		Where(app.AppInstallConfigSync{ID: syncID, AppID: appID, OrgID: org.ID}).
		Preload("Workflow").
		Preload("Workflow.Steps").
		Preload("Workflow.Steps.Approval").
		Preload("Workflow.Steps.Approval.Response").
		Preload("Workflow.StepGroups").
		Preload("InstallConfigSyncs").
		Preload("InstallConfigSyncs.Versions").
		Preload("InstallCreationApproval").
		Preload("VCSConnectionCommit").
		First(&sync).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install sync: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, sync)
}
