package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminWorkflowStepApproveRequest struct {
	StepID string `json:"step_id"`
}

// @ID						AdminWorkflowStepApprove
// @Description.markdown	update_install_runner.md
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminWorkflowStepApproveRequest	true	"Input"
// @Produce					json
// @Success					200	{object}	app.WorkflowStepApprovalResponse
// @Router					/v1/admin-workflow-step-approve [post]
func (s *service) AdminWorkflowStepApprove(ctx *gin.Context) {
	var req AdminWorkflowStepApproveRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	approvalResponse, err := s.approveWorkflowStep(ctx, req.StepID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, approvalResponse)

}

func (s *service) approveWorkflowStep(ctx *gin.Context, stepID string) (*app.WorkflowStepApprovalResponse, error) {
	var installWorkflowStep *app.WorkflowStep
	res := s.db.WithContext(ctx).
		Where("id = ?", stepID).
		Preload("Approval").
		Preload("Approval.Response").
		Preload("PolicyValidation").
		First(&installWorkflowStep)
	if res.Error != nil {
		return nil, errors.Wrapf(res.Error, "unable to find install workflow step with ID: %s", stepID)

	}

	if installWorkflowStep.Approval == nil {
		return nil, fmt.Errorf("install workflow step with ID: %s does not have an approval", stepID)
	}

	if installWorkflowStep.Approval.Response != nil {
		return nil, fmt.Errorf("install workflow step with ID: %s already has an approval response", stepID)
	}

	response := &app.WorkflowStepApprovalResponse{
		OrgID:                         installWorkflowStep.OrgID,
		InstallWorkflowStepApprovalID: installWorkflowStep.Approval.ID,
		Type:                          app.WorkflowStepApprovalResponseTypeApprove,
		Note:                          "Admin approved the step",
	}

	res = s.db.WithContext(ctx).Create(&response)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create response: %w", res.Error)
	}

	return response, nil
}

// TODO: Remove. Deprecated.
// @ID						AdminInstallWorkflowStepApprove
// @Description.markdown	update_install_runner.md
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminWorkflowStepApproveRequest	true	"Input"
// @Produce					json
// @Success					200	{object}	app.WorkflowStepApprovalResponse
// @Router					/v1/admin-install-workflow-step-approve [post]
func (s *service) AdminInstallWorkflowStepApprove(ctx *gin.Context) {
	var req AdminWorkflowStepApproveRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	approvalResponse, err := s.approveWorkflowStep(ctx, req.StepID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, approvalResponse)
}
