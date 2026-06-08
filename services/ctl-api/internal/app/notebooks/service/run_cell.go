package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	notebookclient "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type RunCellRequest struct {
	// IdempotencyKey deduplicates retried run requests. Optional; a server-side
	// key is generated when empty.
	IdempotencyKey string `json:"idempotency_key" validate:"omitempty,max=255"`
}

// @ID				RunNotebookCell
// @Summary		run a notebook cell on the install's runner
// @Description	dispatches the cell to the notebook's warm Temporal workflow and records a NotebookCellRun linking to the underlying execution + log stream. Returns once the run is queued, not when it finishes.
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string			true	"install ID"
// @Param			notebook_id	path	string			true	"notebook ID"
// @Param			cell_id		path	string			true	"cell ID"
// @Param			req			body	RunCellRequest	false	"Input"
// @Success		202			{object}	app.NotebookCellRun
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/cells/{cell_id}/runs [post]
func (s *service) RunCell(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	cell, err := s.getCell(ctx, org.ID, install.ID, ctx.Param("notebook_id"), ctx.Param("cell_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	var req RunCellRequest
	// body is optional
	_ = ctx.ShouldBindJSON(&req)
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.NewString()
	}

	account, err := cctx.AccountFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	resp, err := s.notebookClient.RunCell(ctx, notebookclient.RunCellRequest{
		NotebookID:     cell.NotebookID,
		CellID:         cell.ID,
		IdempotencyKey: req.IdempotencyKey,
		OrgID:          org.ID,
		InstallID:      install.ID,
		TriggeredByID:  account.ID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to run cell: %w", err))
		return
	}

	// Reload the cell run (created by the workflow) so the response carries the
	// full record, including the log stream once the workflow records it.
	var run app.NotebookCellRun
	if res := s.db.WithContext(ctx).
		First(&run, "id = ?", resp.NotebookCellRunID); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to load cell run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusAccepted, run)
}
