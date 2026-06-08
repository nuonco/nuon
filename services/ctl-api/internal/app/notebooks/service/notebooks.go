package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	notebookstart "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/signals/start"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// notebookQueueName is the name of the per-notebook queue that owns the warm
// workflow lifecycle. The notebook workflow runs in the installs namespace.
const (
	notebookQueueName      = "notebook"
	notebookQueueNamespace = "installs"
)

type CreateNotebookRequest struct {
	Name        string `json:"name" validate:"max=255"`
	Description string `json:"description" validate:"max=2000"`
}

// @ID				CreateNotebook
// @Summary		create a notebook for an install
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string					true	"install ID"
// @Param			req			body	CreateNotebookRequest	true	"Input"
// @Success		201			{object}	app.Notebook
// @Router			/v1/installs/{install_id}/notebooks [post]
func (s *service) CreateNotebook(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateNotebookRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	nb := app.Notebook{
		OrgID:       org.ID,
		InstallID:   install.ID,
		Name:        req.Name,
		Description: req.Description,
		Status:      app.NotebookStatusActive,
	}
	if res := s.db.WithContext(ctx).Create(&nb); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create notebook: %w", res.Error))
		return
	}

	// Bring the notebook's warm workflow online via its queue. Non-blocking:
	// if this fails the notebook is still usable — the first cell run lazily
	// starts the workflow via update-with-start.
	if err := s.startNotebookWorkflow(ctx, &nb); err != nil {
		s.l.Warn("unable to start notebook workflow via queue",
			zap.String("notebook-id", nb.ID), zap.Error(err))
	}

	ctx.JSON(http.StatusCreated, nb)
}

// startNotebookWorkflow ensures the notebook owns a queue and enqueues a
// notebook-start signal so the warm per-notebook workflow comes online (and can
// be re-dispatched for recovery). Cell runs dispatch to that workflow directly.
func (s *service) startNotebookWorkflow(ctx context.Context, nb *app.Notebook) error {
	ownerType := plugins.TableName(s.db, app.Notebook{})

	q, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OrgID:       &nb.OrgID,
		OwnerID:     nb.ID,
		OwnerType:   ownerType,
		Namespace:   notebookQueueNamespace,
		Name:        notebookQueueName,
		MaxInFlight: 1,
		MaxDepth:    10,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure notebook queue: %w", err)
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   nb.ID,
		OwnerType: ownerType,
		Signal:    &notebookstart.Signal{NotebookID: nb.ID},
	}); err != nil {
		return fmt.Errorf("unable to enqueue notebook-start signal: %w", err)
	}

	return nil
}

// @ID				GetNotebooks
// @Summary		list notebooks for an install
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Success		200			{array}	app.Notebook
// @Router			/v1/installs/{install_id}/notebooks [get]
func (s *service) GetNotebooks(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	notebooks := []*app.Notebook{}
	res := s.db.WithContext(ctx).
		Where(app.Notebook{OrgID: org.ID, InstallID: install.ID}).
		Order("updated_at DESC").
		Find(&notebooks)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list notebooks: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, notebooks)
}

// @ID				GetNotebook
// @Summary		get a notebook with its ordered cells and each cell's latest run
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			notebook_id	path	string	true	"notebook ID"
// @Success		200			{object}	app.Notebook
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id} [get]
func (s *service) GetNotebook(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var nb app.Notebook
	res := s.db.WithContext(ctx).
		Preload("Cells", func(db *gorm.DB) *gorm.DB {
			return db.Order("notebook_cells.position ASC")
		}).
		Where(app.Notebook{OrgID: org.ID, InstallID: install.ID}).
		First(&nb, "id = ?", ctx.Param("notebook_id"))
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get notebook: %w", res.Error))
		return
	}

	s.attachLatestRuns(ctx, org.ID, &nb)

	ctx.JSON(http.StatusOK, nb)
}

type UpdateNotebookRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	Status      *string `json:"status" validate:"omitempty,oneof=active archived"`
}

// @ID				UpdateNotebook
// @Summary		update a notebook
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string					true	"install ID"
// @Param			notebook_id	path	string					true	"notebook ID"
// @Param			req			body	UpdateNotebookRequest	true	"Input"
// @Success		200			{object}	app.Notebook
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id} [patch]
func (s *service) UpdateNotebook(ctx *gin.Context) {
	org, install, err := s.gate(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateNotebookRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	nb, err := s.getNotebook(ctx, org.ID, install.ID, ctx.Param("notebook_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = app.NotebookStatus(*req.Status)
	}
	if len(updates) > 0 {
		if res := s.db.WithContext(ctx).Model(nb).Updates(updates); res.Error != nil {
			ctx.Error(fmt.Errorf("unable to update notebook: %w", res.Error))
			return
		}
	}

	ctx.JSON(http.StatusOK, nb)
}

// @ID				DeleteNotebook
// @Summary		delete a notebook
// @Tags			notebooks
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Param			install_id	path	string	true	"install ID"
// @Param			notebook_id	path	string	true	"notebook ID"
// @Success		204
// @Router			/v1/installs/{install_id}/notebooks/{notebook_id} [delete]
func (s *service) DeleteNotebook(ctx *gin.Context) {
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

	if res := s.db.WithContext(ctx).Delete(nb); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete notebook: %w", res.Error))
		return
	}

	ctx.Status(http.StatusNoContent)
}

// attachLatestRuns populates each cell's LatestRun with its most recent run.
func (s *service) attachLatestRuns(ctx *gin.Context, orgID string, nb *app.Notebook) {
	for i := range nb.Cells {
		var run app.NotebookCellRun
		res := s.db.WithContext(ctx).
			Where(app.NotebookCellRun{OrgID: orgID, CellID: nb.Cells[i].ID}).
			Order("created_at DESC").
			First(&run)
		if res.Error == nil {
			nb.Cells[i].LatestRun = &run
		}
	}
}
