package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetWorkflowStep
// @Summary					get a workflow step
// @Description.markdown	get_workflow_step.md
// @Param					workflow_id		path	string	true	"workflow id"
// @Param					step_id		path	string	true	"step id"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.WorkflowStep
// @Router					/v1/workflows/{workflow_id}/steps/{step_id} [GET]
func (s *service) GetWorkflowStep(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")

	workflow, err := s.getWorkflowStep(ctx, org.ID, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step"))
		return
	}

	ctx.JSON(http.StatusOK, workflow)
}

// TODO: Remove. Deprecated.
// @ID						GetInstallWorkflowStep
// @Summary					get an install workflow step
// @Description.markdown	get_workflow_step.md
// @Param					install_workflow_id		path	string	true	"workflow id"
// @Param					install_workflow_step_id		path	string	true	"step id"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.WorkflowStep
// @Router					/v1/install-workflows/{install_workflow_id}/steps/{install_workflow_step_id} [GET]
// @Deprecated
func (s *service) GetInstallWorkflowStep(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("install_workflow_id")
	stepID := ctx.Param("install_workflow_step_id")

	installWorkflow, err := s.getWorkflowStep(ctx, org.ID, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow step"))
		return
	}

	ctx.JSON(http.StatusOK, installWorkflow)
}

func (s *service) getWorkflowStep(ctx *gin.Context, orgID, workflowID, stepID string) (*app.WorkflowStep, error) {
	var installWorkflowStep app.WorkflowStep
	res := s.db.WithContext(ctx).
		Where("id = ? AND install_workflow_id = ? AND org_id = ?", stepID, workflowID, orgID).
		Preload("CreatedBy").
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Approval.Response").
		Preload("PolicyValidation").
		First(&installWorkflowStep)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow step")
	}

	return &installWorkflowStep, nil
}
