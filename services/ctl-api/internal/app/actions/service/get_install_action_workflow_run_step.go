package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetInstallActionRunStep
// @Summary				get action workflow run step by install id and step id
// @Description.markdown	get_install_action_workflow_run_step.md
// @Param					install_id		path	string	true	"install ID"
// @Param					run_id	path	string	true	"workflow run ID"
// @Param					step_id			path	string	true	"step ID"
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallActionWorkflowRunStep
// @Router					/v1/installs/{install_id}/actions/runs/{run_id}/steps/{step_id} [get]
func (s *service) GetInstallActionRunStep(ctx *gin.Context) {
	workflowRunID := ctx.Param("run_id")
	stepID := ctx.Param("step_id")
	installID := ctx.Param("install_id")

	step, err := s.getInstallActionWorkflowRunStep(ctx, installID, workflowRunID, stepID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, step)
}

// @ID						GetInstallActionWorkflowRunStep
// @Summary				get action workflow run step by install id and step id
// @Description.markdown	get_install_action_workflow_run_step.md
// @Param					install_id		path	string	true	"install ID"
// @Param					workflow_run_id	path	string	true	"workflow run ID"
// @Param					step_id			path	string	true	"step ID"
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallActionWorkflowRunStep
// @Router					/v1/installs/{install_id}/action-workflows/runs/{workflow_run_id}/steps/{step_id} [get]
func (s *service) GetInstallActionWorkflowRunStep(ctx *gin.Context) {
	workflowRunID := ctx.Param("workflow_run_id")
	stepID := ctx.Param("step_id")
	installID := ctx.Param("install_id")

	step, err := s.getInstallActionWorkflowRunStep(ctx, installID, workflowRunID, stepID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, step)
}

func (s *service) getInstallActionWorkflowRunStep(ctx context.Context, installID, workflowRunID, stepID string) (*app.InstallActionWorkflowRunStep, error) {
	var step app.InstallActionWorkflowRunStep
	res := s.db.WithContext(ctx).
		Joins("JOIN install_action_workflow_runs ON install_action_workflow_runs.id=install_action_workflow_run_steps.workflow_run_id").
		Where("install_action_workflow_runs.id = ?", workflowRunID).
		Where("install_action_workflow_runs.install_id = ?", installID).
		Where("install_action_workflow_run_steps.id = ?", stepID).
		First(&step)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find step")
	}

	return &step, nil
}
