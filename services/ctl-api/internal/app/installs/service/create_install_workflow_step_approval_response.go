package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateInstallWorkflowStepApprovalResponseRequest struct {
	ResponseType app.InstallWorkflowStepResponseType `json:"response_type"`
	Note         string                              `json:"note"`
}

func (c *CreateInstallWorkflowStepApprovalResponseRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateInstallWorkflowStepApprovalResponse
// @Summary				deploy a build to an install
// @Description.markdown	create_install_workflow_step_approval_response.md
// @Param	install_workflow_id		path	string	true	"workflow id"
// @Param	install_workflow_step_id		path	string	true	"step id"
// @Param	approval_id					path	string	true	"approval id"
// @Param					req			body	CreateInstallWorkflowStepApprovalResponseRequest	true	"Input"
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
// @Success				201	{object}	app.InstallWorkflowStepApprovalResponse
// @Router			/v1/install-workflows/{install_workflow_id}/steps/{install_workflow_step_id}/approvals/{approval_id}/response [post]
func (s *service) CreateInstallWorkflowStepApprovalResponse(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	var req CreateInstallWorkflowStepApprovalResponseRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	workflowID := ctx.Param("install_workflow_id")
	stepID := ctx.Param("install_workflow_step_id")
	approvalID := ctx.Param("approval_id")

	_, err = s.getInstallWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow step"))
		return
	}

	approval, err := s.getInstallWorkflowStepApproval(ctx, org.ID, approvalID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow step"))
		return
	}

	if approval.Response != nil {
		ctx.Error(fmt.Errorf("install workflow step approval already has a response"))
		return
	}

	response, err := s.createInstallWorkflowStepApprovalResponse(ctx, approval.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	// TODO: we still need to trigger some background job to process the approval

	ctx.JSON(http.StatusCreated, response)
}

func (s *service) createInstallWorkflowStepApprovalResponse(ctx *gin.Context, approvalID string, req *CreateInstallWorkflowStepApprovalResponseRequest) (*app.InstallWorkflowStepApprovalResponse, error) {
	response := app.InstallWorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approvalID,
		Type:                          req.ResponseType,
		Note:                          req.Note,
	}

	res := s.db.WithContext(ctx).Create(&response)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install deploy: %w", res.Error)
	}

	return &response, nil
}
