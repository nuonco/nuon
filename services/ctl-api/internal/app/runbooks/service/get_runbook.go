package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID				GetRunbook
// @Summary		get a runbook
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			app_id		path	string	true	"app ID"
// @Param			runbook_id	path	string	true	"runbook ID"
// @Success		200			{object}	app.Runbook
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/apps/{app_id}/runbooks/{runbook_id} [get]
func (s *service) GetRunbook(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	runbookIDOrName := ctx.Param("runbook_id")
	appID := ctx.Param("app_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var runbook app.Runbook
	res := s.db.WithContext(ctx).
		Preload("Configs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at DESC").Limit(1)
		}).
		Preload("Configs.Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Configs.Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.Runbook{OrgID: org.ID, AppID: appID}).
		Where("id = ? OR name = ?", runbookIDOrName, runbookIDOrName).
		First(&runbook)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbook: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, runbook)
}
