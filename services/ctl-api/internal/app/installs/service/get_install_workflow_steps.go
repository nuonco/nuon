package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetInstallWorkflowSteps
// @Summary				get an install workflow
// @Description.markdown	get_install_workflow_steps.md
// @Param install_workflow_id	path	string true "install workflow ID"
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
// @Success				200	{array}		app.InstallWorkflowStep
// @Router					/v1/install-workflows/{install_workflow_id}/steps [GET]
func (s *service) GetInstallWorkflowSteps(ctx *gin.Context) {
	workflowID := ctx.Param("install_workflow_id")

	steps, err := s.getInstallWorkflowSteps(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow steps"))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

func (s *service) getInstallWorkflowSteps(ctx *gin.Context, workflowID string) ([]app.InstallWorkflowStep, error) {
	var steps []app.InstallWorkflowStep

	res := s.db.WithContext(ctx).
		Where(app.InstallWorkflowStep{
			InstallWorkflowID: workflowID,
		}).
		Preload("Approval").
		Preload("PolicyValidation").
		Order("idx ASC").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	return steps, nil
}
