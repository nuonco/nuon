package service

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

const (
	defaultCellTimeoutSeconds = 300
	maxCellTimeoutSeconds     = 3600
)

type CreateCellRequest struct {
	Name             string            `json:"name" validate:"max=255"`
	InlineContents   string            `json:"inline_contents" validate:"required_without=Command"`
	Command          string            `json:"command" validate:"required_without=InlineContents"`
	EnvVars          map[string]string `json:"env_vars"`
	Timeout          int               `json:"timeout,omitempty" validate:"omitempty,min=1,max=3600"`
	Role             string            `json:"role"`
	EnableKubeConfig *bool             `json:"enable_kube_config" extensions:"x-nullable"`
}

func (r *CreateCellRequest) validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return err
	}
	if r.InlineContents != "" && r.Command != "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("provide either inline_contents or command, not both"),
			Description: "invalid request input",
		}
	}
	if r.Timeout == 0 {
		r.Timeout = defaultCellTimeoutSeconds
	}
	return nil
}

// @ID				CreateNotebookCell
// @Summary		add a cell to a notebook
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string				true	"install ID"
// @Param			notebook_id	path	string				true	"notebook ID"
// @Param			req			body	CreateCellRequest	true	"Input"
// @Success		201			{object}	app.NotebookCell
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/cells [post]
func (s *service) CreateCell(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	nb, err := s.getNotebook(ctx, org.ID, install.ID, ctx.Param("notebook_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateCellRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := req.validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// new cell goes to the end of the notebook
	var maxPos *int
	s.db.WithContext(ctx).
		Model(&app.NotebookCell{}).
		Where(app.NotebookCell{NotebookID: nb.ID}).
		Select("MAX(position)").
		Scan(&maxPos)
	position := 0
	if maxPos != nil {
		position = *maxPos + 1
	}

	cell := app.NotebookCell{
		OrgID:            org.ID,
		NotebookID:       nb.ID,
		Position:         position,
		Revision:         1,
		Name:             req.Name,
		InlineContents:   req.InlineContents,
		Command:          req.Command,
		EnvVars:          dbgenerics.ToHstore(req.EnvVars),
		Timeout:          time.Duration(req.Timeout) * time.Second,
		Role:             req.Role,
		EnableKubeConfig: enableKubeConfig(req.EnableKubeConfig),
	}
	if res := s.db.WithContext(ctx).Create(&cell); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create cell: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, cell)
}

type UpdateCellRequest struct {
	Name             *string            `json:"name" validate:"omitempty,max=255"`
	InlineContents   *string            `json:"inline_contents"`
	Command          *string            `json:"command"`
	EnvVars          *map[string]string `json:"env_vars"`
	Timeout          *int               `json:"timeout" validate:"omitempty,min=1,max=3600"`
	Role             *string            `json:"role"`
	EnableKubeConfig *bool              `json:"enable_kube_config" extensions:"x-nullable"`
}

// @ID				UpdateNotebookCell
// @Summary		edit a cell (bumps its revision)
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string				true	"install ID"
// @Param			notebook_id	path	string				true	"notebook ID"
// @Param			cell_id		path	string				true	"cell ID"
// @Param			req			body	UpdateCellRequest	true	"Input"
// @Success		200			{object}	app.NotebookCell
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/cells/{cell_id} [patch]
func (s *service) UpdateCell(ctx *gin.Context) {
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

	var req UpdateCellRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.InlineContents != nil {
		updates["inline_contents"] = *req.InlineContents
	}
	if req.Command != nil {
		updates["command"] = *req.Command
	}
	if req.EnvVars != nil {
		updates["env_vars"] = dbgenerics.ToHstore(*req.EnvVars)
	}
	if req.Timeout != nil {
		updates["timeout"] = time.Duration(*req.Timeout) * time.Second
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.EnableKubeConfig != nil {
		updates["enable_kube_config"] = enableKubeConfig(req.EnableKubeConfig)
	}

	if len(updates) > 0 {
		// any edit bumps the revision so the UI can flag "edited since last run"
		updates["revision"] = cell.Revision + 1
		if res := s.db.WithContext(ctx).Model(cell).Updates(updates); res.Error != nil {
			ctx.Error(fmt.Errorf("unable to update cell: %w", res.Error))
			return
		}
	}

	ctx.JSON(http.StatusOK, cell)
}

// @ID				DeleteNotebookCell
// @Summary		delete a cell
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			notebook_id	path	string	true	"notebook ID"
// @Param			cell_id		path	string	true	"cell ID"
// @Success		204
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/cells/{cell_id} [delete]
func (s *service) DeleteCell(ctx *gin.Context) {
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

	if res := s.db.WithContext(ctx).Delete(cell); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete cell: %w", res.Error))
		return
	}

	ctx.Status(http.StatusNoContent)
}

type ReorderCellsRequest struct {
	CellIDs []string `json:"cell_ids" validate:"required,min=1"`
}

// @ID				ReorderNotebookCells
// @Summary		reorder a notebook's cells
// @Description	accepts the full ordered list of cell IDs and assigns positions
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string				true	"install ID"
// @Param			notebook_id	path	string				true	"notebook ID"
// @Param			req			body	ReorderCellsRequest	true	"Input"
// @Success		200			{object}	app.Notebook
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id}/cells/reorder [put]
func (s *service) ReorderCells(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	nb, err := s.getNotebook(ctx, org.ID, install.ID, ctx.Param("notebook_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	var req ReorderCellsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for pos, cellID := range req.CellIDs {
			res := tx.Model(&app.NotebookCell{}).
				Where(app.NotebookCell{OrgID: org.ID, NotebookID: nb.ID}).
				Where("id = ?", cellID).
				Update("position", pos)
			if res.Error != nil {
				return res.Error
			}
		}
		return nil
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reorder cells: %w", err))
		return
	}

	ctx.Status(http.StatusOK)
}

func (s *service) getCell(ctx *gin.Context, orgID, installID, notebookID, cellID string) (*app.NotebookCell, error) {
	// confirm the notebook belongs to the org+install before touching cells
	if _, err := s.getNotebook(ctx, orgID, installID, notebookID); err != nil {
		return nil, err
	}

	var cell app.NotebookCell
	res := s.db.WithContext(ctx).
		Where(app.NotebookCell{OrgID: orgID, NotebookID: notebookID}).
		First(&cell, "id = ?", cellID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get cell: %w", res.Error)
	}
	return &cell, nil
}

func enableKubeConfig(v *bool) sql.NullBool {
	if v != nil {
		return generics.NewNullBoolFromPtr(v)
	}
	enabled := true
	return generics.NewNullBoolFromPtr(&enabled)
}
