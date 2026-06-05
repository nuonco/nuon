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

// @ID				GetInstallRunbookRuns
// @Summary		get runbook runs for an install
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			runbook_id	query	string	false	"filter by runbook ID or name"
// @Param			offset		query	int		false	"offset"	Default(0)
// @Param			limit		query	int		false	"limit"		Default(10)
// @Success		200			{array}	app.InstallRunbookRun
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/installs/{install_id}/runbook-runs [get]
func (s *service) GetInstallRunbookRuns(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	installID := ctx.Param("install_id")
	runbookIDOrName := ctx.Query("runbook_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	query := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("InstallRunbook").
		Preload("InstallRunbook.Runbook").
		Preload("InstallWorkflow").
		Where(app.InstallRunbookRun{OrgID: org.ID, InstallID: installID})

	if runbookIDOrName != "" {
		query = query.
			Joins("JOIN install_runbooks ON install_runbooks.id = install_runbook_runs.install_runbook_id AND install_runbooks.deleted_at = 0").
			Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
			Where("install_runbooks.runbook_id = ? OR runbooks.name = ?", runbookIDOrName, runbookIDOrName)
	}

	runs := []*app.InstallRunbookRun{}
	res := query.
		Order("install_runbook_runs.created_at DESC").
		Find(&runs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbook runs: %w", res.Error))
		return
	}

	runs, err = db.HandlePaginatedResponse(ctx, runs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runs)
}

// @ID				GetInstallRunbookRun
// @Summary		get a runbook run
// @Tags			runbooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			run_id		path	string	true	"run ID"
// @Success		200			{object}	app.InstallRunbookRun
// @Failure		400			{object}	stderr.ErrResponse
// @Failure		401			{object}	stderr.ErrResponse
// @Failure		403			{object}	stderr.ErrResponse
// @Failure		404			{object}	stderr.ErrResponse
// @Failure		500			{object}	stderr.ErrResponse
// @Router			/v1/installs/{install_id}/runbook-runs/{run_id} [get]
func (s *service) GetInstallRunbookRun(ctx *gin.Context) {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunbooks)
	if err != nil || !enabled {
		ctx.Error(fmt.Errorf("runbooks feature is not enabled"))
		return
	}

	runID := ctx.Param("run_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var run app.InstallRunbookRun
	res := s.db.WithContext(ctx).
		Preload("InstallRunbook").
		Preload("InstallRunbook.Runbook").
		Preload("RunbookConfig").
		Preload("RunbookConfig.Steps").
		Preload("InstallWorkflow").
		Where(app.InstallRunbookRun{OrgID: org.ID}).
		First(&run, "id = ?", runID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get runbook run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, run)
}
