package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateWorkflowStepApprovalResponseRequest struct {
	ResponseType app.WorkflowStepResponseType `json:"response_type"`
	Note         string                       `json:"note"`
}

func (c *CreateWorkflowStepApprovalResponseRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// NOTE(fd): y tho? because if we return app.WorkflowStepApprovalResponse it breaks some naive
// SDK generators (cough, openapi-python)
type CreateWorkflowStepApprovalResponseResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Note string `json:"note"`
}

// @ID						CreateWorkflowStepApprovalResponse
// @Summary					Create an approval response for a workflow step.
// @Description.markdown	create_workflow_step_approval_response.md
// @Param					workflow_id			path	string	true	"workflow id"
// @Param					step_id	path	string	true	"step id"
// @Param					approval_id			path	string	true	"approval id"
// @Param					req					body	CreateWorkflowStepApprovalResponseRequest	true	"Input"
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
// @Success					201	{object}	CreateWorkflowStepApprovalResponseResponse
// @Router					/v1/workflows/{workflow_id}/steps/{step_id}/approvals/{approval_id}/response [post]
func (s *service) CreateWorkflowStepApprovalResponse(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	var req CreateWorkflowStepApprovalResponseRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")
	approvalID := ctx.Param("approval_id")

	_, err = s.getWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step"))
		return
	}

	approval, err := s.getWorkflowStepApproval(ctx, org.ID, approvalID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow step"))
		return
	}

	if approval.Response != nil {
		ctx.Error(fmt.Errorf("workflow step approval already has a response"))
		return
	}

	// create the response
	wfsaResponse, err := s.createWorkflowStepApprovalResponse(ctx, approval.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	response := CreateWorkflowStepApprovalResponseResponse{
		ID:   wfsaResponse.ID,
		Type: string(wfsaResponse.Type),
		Note: string(wfsaResponse.Note),
	}

	ctx.JSON(http.StatusCreated, response)
}

func (s *service) createWorkflowStepApprovalResponse(ctx *gin.Context, approvalID string, req *CreateWorkflowStepApprovalResponseRequest) (*app.WorkflowStepApprovalResponse, error) {
	response := app.WorkflowStepApprovalResponse{
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
