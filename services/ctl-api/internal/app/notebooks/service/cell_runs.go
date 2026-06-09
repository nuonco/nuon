package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID				GetNotebookCellRuns
// @Summary		list a cell's run history (newest first)
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			notebook_id	path	string	true	"notebook ID"
// @Param			cell_id		path	string	true	"cell ID"
// @Param			offset		query	int		false	"offset"	Default(0)
// @Param			limit		query	int		false	"limit"		Default(20)
// @Success		200			{array}	app.NotebookCellRun
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/cells/{cell_id}/runs [get]
func (s *service) GetCellRuns(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if _, err := s.getCell(ctx, org.ID, install.ID, ctx.Param("notebook_id"), ctx.Param("cell_id")); err != nil {
		ctx.Error(err)
		return
	}

	runs := []*app.NotebookCellRun{}
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where(app.NotebookCellRun{OrgID: org.ID, CellID: ctx.Param("cell_id")}).
		Order("created_at DESC").
		Find(&runs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list cell runs: %w", res.Error))
		return
	}

	runs, err = db.HandlePaginatedResponse(ctx, runs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, runs)
}

// @ID				GetNotebookCellRun
// @Summary		get a single cell run (includes log_stream_id for tailing)
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			notebook_id	path	string	true	"notebook ID"
// @Param			run_id		path	string	true	"run ID"
// @Success		200			{object}	app.NotebookCellRun
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/runs/{run_id} [get]
func (s *service) GetCellRun(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var run app.NotebookCellRun
	res := s.db.WithContext(ctx).
		Where(app.NotebookCellRun{OrgID: org.ID, InstallID: install.ID, NotebookID: ctx.Param("notebook_id")}).
		First(&run, "id = ?", ctx.Param("run_id"))
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get cell run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, run)
}
