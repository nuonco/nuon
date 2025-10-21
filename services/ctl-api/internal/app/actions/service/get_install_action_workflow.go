package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetInstallAction
// @Summary				get an install action
// @Description.markdown	get_install_action_workflow.md
// @Param					install_id			path	string	true	"install ID"
// @Param					action_id	path	string	true	"action ID"
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
// @Success				200	{object}	app.InstallActionWorkflow
// @Router					/v1/installs/{install_id}/actions/{action_id} [get]
func (s *service) GetInstallAction(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	actionWorkflowID := ctx.Param("action_id")

	installActionWorkflow, err := s.getInstallActionWorkflow(ctx, installID, actionWorkflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install action workflow"))
		return
	}

	ctx.JSON(http.StatusOK, installActionWorkflow)
}

// @ID						GetInstallActionWorkflow
// @Summary				get an install action workflow
// @Description.markdown	get_install_action_workflow.md
// @Param					install_id			path	string	true	"install ID"
// @Param					action_workflow_id	path	string	true	"workflow ID"
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
// @Success				200	{object}	app.InstallActionWorkflow
// @Router					/v1/installs/{install_id}/action-workflows/{action_workflow_id} [get]
func (s *service) GetInstallActionWorkflow(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	actionWorkflowID := ctx.Param("action_workflow_id")

	installActionWorkflow, err := s.getInstallActionWorkflow(ctx, installID, actionWorkflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install action workflow"))
		return
	}

	ctx.JSON(http.StatusOK, installActionWorkflow)
}

func (s *service) getInstallActionWorkflow(ctx context.Context, installID, actionWorkflowID string) (*app.InstallActionWorkflow, error) {
	var installActionWorkflow app.InstallActionWorkflow
	res := s.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{
			InstallID:        installID,
			ActionWorkflowID: actionWorkflowID,
		}).
		Preload("ActionWorkflow").
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Limit(5).Order("install_action_workflow_runs.created_at DESC")
		}).
		First(&installActionWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow")
	}

	return &installActionWorkflow, nil
}
