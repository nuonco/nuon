package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetInstallWorkflowStep
// @Summary				get an install workflow step
// @Description.markdown	get_install_workflow_step.md
// @Param	install_workflow_id		path	string	true	"workflow id"
// @Param	install_workflow_step_id		path	string	true	"step id"
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
// @Success				200	{array}		app.InstallSandboxRun
// @Router /v1/install-workflows/{install_workflow_id}/steps/{install_workflow_step_id} [GET]
func (s *service) GetInstallWorkflowStep(ctx *gin.Context) {
	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")

	installWorkflow, err := s.getInstallWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow step"))
		return
	}

	ctx.JSON(http.StatusOK, installWorkflow)
}

func (s *service) getInstallWorkflowStep(ctx *gin.Context, workflowID, stepID string) (*app.InstallWorkflowStep, error) {
	var installWorkflowStep app.InstallWorkflowStep
	res := s.db.WithContext(ctx).
		Where(app.InstallWorkflowStep{
			ID: workflowID,
		}).
		Preload("Approval").
		Preload("PolicyValidation").
		First(&installWorkflowStep)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	return &installWorkflowStep, nil
}
