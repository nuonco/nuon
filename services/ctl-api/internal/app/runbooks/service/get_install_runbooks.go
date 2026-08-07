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

const currentAppConfigRunbookSubquery = `
	SELECT 1 FROM runbook_configs rc
	JOIN installs i ON i.id = install_runbooks.install_id
	WHERE rc.runbook_id = install_runbooks.runbook_id
		AND rc.app_config_id = i.app_config_id
		AND rc.deleted_at = 0
`

// appConfigRunbookFilter keeps only runbooks present in the install's current app config
// (synced) when syncedOnly is true, or only runbooks no longer in the current app config
// when syncedOnly is false.
func appConfigRunbookFilter(syncedOnly bool) string {
	if syncedOnly {
		return "EXISTS (" + currentAppConfigRunbookSubquery + ")"
	}
	return "NOT EXISTS (" + currentAppConfigRunbookSubquery + ")"
}

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
// @Param			q			query	string	false	"search by runbook name or ID"
// @Param			synced		query	bool	false	"return runbooks in the install's current app config; set false to return only runbooks no longer in it"	Default(true)
// @Success		200			{array}	app.InstallRunbook
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/installs/{install_id}/runbooks [get]
func (s *service) GetInstallRunbooks(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	q := ctx.Query("q")
	synced := ctx.Query("synced") != "false"

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
		Preload("Runbook").
		// Pinned config, not the app's newest: runs execute the pinned one, so any
		// other version hands callers step IDs it will reject. Unique on
		// (runbook_id, app_config_id), so this yields at most one per runbook.
		Preload("Runbook.Configs", func(tx *gorm.DB) *gorm.DB {
			return tx.Where(app.RunbookConfig{AppConfigID: install.AppConfigID})
		}).
		Preload("Runbook.Configs.Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Runbook.Configs.Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Runs", func(tx *gorm.DB) *gorm.DB {
			return tx.Scopes(scopes.WithOverrideTable("install_runbook_runs_latest_view_v1"))
		}).
		Where(app.InstallRunbook{OrgID: org.ID, InstallID: installID}).
		Where(appConfigRunbookFilter(synced)).
		Order("install_runbooks.created_at DESC")

	if q != "" {
		tx = tx.Where("runbooks.name ILIKE ? OR install_runbooks.id = ?", "%"+q+"%", q)
	}

	installRunbooks := []*app.InstallRunbook{}
	res := tx.Find(&installRunbooks)
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
