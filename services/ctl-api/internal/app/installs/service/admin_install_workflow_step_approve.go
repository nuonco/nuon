package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminInstallWorkflowStepApproveRequest struct {
	StepID string `json:"step_id"`
}

// @ID						AdminInstallWorkflowStepApprove
// @Description.markdown	update_install_runner.md
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminInstallWorkflowStepApproveRequest	true	"Input"
// @Produce				json
// @Success				200	{object}	app.InstallWorkflowStepApprovalResponse
// @Router					/v1/admin-install-workflow-step-approve [post]
func (s *service) AdminInstallWorkflowStepApprove(ctx *gin.Context) {
	var req AdminInstallWorkflowStepApproveRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	var installWorkflowStep app.InstallWorkflowStep
	res := s.db.WithContext(ctx).
		Where("id = ?", req.StepID).
		Preload("Approval").
		Preload("Approval.Response").
		Preload("PolicyValidation").
		First(&installWorkflowStep)
	if res.Error != nil {
		ctx.Error(errors.Wrapf(res.Error, "unable to find install workflow step with ID: %s", req.StepID))
		return

	}

	if installWorkflowStep.Approval == nil {
		ctx.Error(fmt.Errorf("install workflow step with ID: %s does not have an approval", req.StepID))
		return
	}

	if installWorkflowStep.Approval.Response != nil {
		ctx.Error(fmt.Errorf("install workflow step with ID: %s already has an approval response", req.StepID))
		return
	}

	response := app.InstallWorkflowStepApprovalResponse{
		OrgID:                         installWorkflowStep.OrgID,
		InstallWorkflowStepApprovalID: installWorkflowStep.Approval.ID,
		Type:                          app.InstallWorkflowStepApprovalResponseTypeApprove,
		Note:                          "Admin approved the step",
	}

	res = s.db.WithContext(ctx).Create(&response)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create response: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, response)
}
