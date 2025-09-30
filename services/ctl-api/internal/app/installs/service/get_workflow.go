package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetWorkflow
// @Summary					get a workflow
// @Description.markdown	get_workflow.md
// @Param					workflow_id path	string	true	"workflow ID"
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
// @Success					200	{object}	app.Workflow
// @Router					/v1/workflows/{workflow_id} [GET]
func (s *service) GetWorkflow(ctx *gin.Context) {
	workflowID := ctx.Param("workflow_id")

	workflow, err := s.getWorkflow(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflows"))
		return
	}

	ctx.JSON(http.StatusOK, workflow)
}

// TODO: Remove. Deprecated.
// @ID						GetInstallWorkflow
// @Summary					get an install workflow
// @Description.markdown	get_workflow.md
// @Param					install_workflow_id path	string	true	"install workflow ID"
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
// @Success					200	{object}	app.Workflow
// @Router					/v1/install-workflows/{install_workflow_id} [GET]
// @Deprecated
func (s *service) GetInstallWorkflow(ctx *gin.Context) {
	workflowID := ctx.Param("install_workflow_id")

	installWorkflow, err := s.getWorkflow(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflows"))
		return
	}

	ctx.JSON(http.StatusOK, installWorkflow)
}

func (s *service) getWorkflow(ctx *gin.Context, workflowID string) (*app.Workflow, error) {
	var installWorkflow app.Workflow
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval").
		Preload("Steps.Approval.Response").
		Where("id = ?", workflowID).
		First(&installWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow")
	}

	return &installWorkflow, nil
}
