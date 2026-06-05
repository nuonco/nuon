package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID				GetInstallRunbooks
// @Summary		get runbooks for an install
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			offset		query	int		false	"offset"	Default(0)
// @Param			limit		query	int		false	"limit"		Default(10)
// @Success		200			{array}	app.InstallRunbook
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/installs/{install_id}/runbooks [get]
func (s *service) GetInstallRunbooks(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	installID := ctx.Param("install_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installRunbooks := []*app.InstallRunbook{}
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
		Preload("Runbook").
		Preload("Runbook.Configs", func(tx *gorm.DB) *gorm.DB {
			return tx.Scopes(scopes.WithOverrideTable("runbook_configs_latest_view_v1"))
		}).
		Preload("Runbook.Configs.Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Runs", func(tx *gorm.DB) *gorm.DB {
			return tx.Scopes(scopes.WithOverrideTable("install_runbook_runs_latest_view_v1"))
		}).
		Where(app.InstallRunbook{OrgID: org.ID, InstallID: installID}).
		Order("install_runbooks.created_at DESC").
		Find(&installRunbooks)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get install runbooks: %w", res.Error))
		return
	}

	installRunbooks, err = db.HandlePaginatedResponse(ctx, installRunbooks)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installRunbooks)
}
