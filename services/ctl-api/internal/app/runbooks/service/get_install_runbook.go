package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID				GetInstallRunbook
// @Summary		get an install runbook
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			runbook_id	path	string	true	"runbook ID"
// @Success		200			{object}	app.InstallRunbook
// @Router			/v1/installs/{install_id}/runbooks/{runbook_id} [get]
func (s *service) GetInstallRunbook(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	installID := ctx.Param("install_id")
	runbookID := ctx.Param("runbook_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var installRunbook app.InstallRunbook
	res := s.db.WithContext(ctx).
		Preload("Runbook").
		Preload("Runbook.Configs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at DESC").Limit(1)
		}).
		Preload("Runbook.Configs.Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Runs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at DESC").Limit(10)
		}).
		Preload("Runs.InstallWorkflow").
		Where(app.InstallRunbook{OrgID: org.ID, InstallID: installID}).
		First(&installRunbook, "runbook_id = ?", runbookID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install runbook: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, installRunbook)
}
