package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/render"
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
// @Param			runbook_id	path	string	true	"runbook ID or name"
// @Success		200			{object}	app.InstallRunbook
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/installs/{install_id}/runbooks/{runbook_id} [get]
func (s *service) GetInstallRunbook(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	installID := ctx.Param("install_id")
	runbookIDOrName := ctx.Param("runbook_id")
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
		Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
		Where(app.InstallRunbook{OrgID: org.ID, InstallID: installID}).
		Where("install_runbooks.runbook_id = ? OR runbooks.name = ?", runbookIDOrName, runbookIDOrName).
		First(&installRunbook)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install runbook: %w", res.Error))
		return
	}

	// Render runbook config readmes with install state, matching the install
	// readme endpoint pattern (get_install_readme.go).
	if installRunbook.Runbook.Configs != nil {
		installState, err := s.installHelpers.GetInstallState(ctx, installID, true, true)
		if err == nil {
			stateMap, err := installState.AsMap()
			if err == nil {
				for i := range installRunbook.Runbook.Configs {
					cfg := &installRunbook.Runbook.Configs[i]
					if cfg.Readme != "" {
						rendered, _, renderErr := render.RenderWithWarnings(cfg.Readme, stateMap)
						if renderErr != nil {
							zap.L().Warn("unable to render runbook readme", zap.Error(renderErr))
							continue
						}
						cfg.Readme = rendered
					}
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, installRunbook)
}
