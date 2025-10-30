package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type UpdateWorkflowRequest struct {
	ApprovalOption *app.InstallApprovalOption `json:"approval_option" validate:"required"`
}

func (c *UpdateWorkflowRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateWorkflow
// @Summary					update a workflow
// @Description.markdown	update_workflow.md
// @Param					workflow_id path	string	true	"workflow ID"
// @Param					req			body	UpdateWorkflowRequest	true	"Input"
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
// @Router					/v1/workflows/{workflow_id}  [PATCH]
func (s *service) UpdateWorkflow(ctx *gin.Context) {
	workflowID := ctx.Param("workflow_id")

	var req UpdateWorkflowRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	workflow, err := s.updateWorkflow(ctx, workflowID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", workflowID, err))
		return
	}

	ctx.JSON(http.StatusOK, workflow)
}

// TODO: Remove. Deprecated.
// @ID						UpdateInstallWorkflow
// @Summary					update an install workflow
// @Description.markdown	update_workflow.md
// @Param					install_workflow_id path	string	true	"install workflow ID"
// @Param					req			body	UpdateWorkflowRequest	true	"Input"
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
// @Router					/v1/install-workflows/{install_workflow_id}  [PATCH]
// @Deprecated
func (s *service) UpdateInstallWorkflow(ctx *gin.Context) {
	installWorkflowID := ctx.Param("install_workflow_id")

	var req UpdateWorkflowRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	installWorkflow, err := s.updateWorkflow(ctx, installWorkflowID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installWorkflowID, err))
		return
	}

	ctx.JSON(http.StatusOK, installWorkflow)
}

func (s *service) updateWorkflow(ctx context.Context, installWorkflowID string, req *UpdateWorkflowRequest) (*app.Workflow, error) {
	currentWorkflow := app.Workflow{
		ID: installWorkflowID,
	}

	res := s.db.WithContext(ctx).
		Model(&currentWorkflow).
		Updates(req)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install workflow: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("install workflow not found: %w", gorm.ErrRecordNotFound)
	}

	return &currentWorkflow, nil
}
