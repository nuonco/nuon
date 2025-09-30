package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID							GetWorkflowSteps
// @Summary						get all of the steps for a given workflow
// @Description.markdown		get_workflow_steps.md
// @Param workflow_id	path	string true "workflow ID"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success						200	{array}		app.WorkflowStep
// @Router						/v1/workflows/{workflow_id}/steps [GET]
func (s *service) GetWorkflowSteps(ctx *gin.Context) {
	workflowID := ctx.Param("workflow_id")

	steps, err := s.getWorkflowSteps(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow steps"))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

// TODO: Remove. Deprecated.
// @ID							GetInstallWorkflowSteps
// @Summary						get all of the steps for a given install workflow
// @Description.markdown		get_workflow_steps.md
// @Param install_workflow_id	path	string true "install workflow ID"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success						200	{array}		app.WorkflowStep
// @Router						/v1/install-workflows/{install_workflow_id}/steps [GET]
// @Deprecated
func (s *service) GetInstallWorkflowSteps(ctx *gin.Context) {
	workflowID := ctx.Param("install_workflow_id")

	steps, err := s.getWorkflowSteps(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow steps"))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

func (s *service) getWorkflowSteps(ctx *gin.Context, workflowID string) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep

	res := s.db.WithContext(ctx).
		Where(app.WorkflowStep{
			InstallWorkflowID: workflowID,
		}).
		Preload("CreatedBy").
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Approval.Response").
		Preload("PolicyValidation").
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	return steps, nil
}
