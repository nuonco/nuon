package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/forgotten"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// DEPRECATED: This endpoint is deprecated and will be removed in a future release.

// @ID						DeleteInstall
// @Summary				delete an install
// @Description.markdown	delete_install.md
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id} [DELETE]
func (s *service) DeleteInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeDeprovision,
		map[string]string{},
		false,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	lp := lifecyclephase.New(lifecyclephase.Deprovisioning, "Tearing down components and cloud resources")
	s.db.WithContext(ctx).Model(&app.Install{ID: install.ID}).Updates(map[string]any{
		"lifecycle_phase": lp,
	})

	workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, workflowsQueueID, &executeflow.Signal{
		WorkflowID: workflow.ID,
	}, workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	if err := s.enqueueInstallSignal(ctx, signalsQueueID, &forgotten.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.WorkflowResponse{WorkflowID: workflow.ID})
}
