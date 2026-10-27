package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/alignment"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type ResolveInstallDriftRequest alignment.ResolveRequest

func (r *ResolveInstallDriftRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID ResolveInstallDrift
// @Summary launch a deployment that resolves drift for one install resource
// @Description Launches a fresh deployment of the selected target for one resource. approved_plan_id explicitly approves the inspected plan for reuse. Without it, only an already explicitly approved plan can be reused. The fresh plan must match and pass policy and safety checks; otherwise it waits for normal approval. Returns an existing open deployment (200) instead of creating a duplicate.
// @Param install_id path string true "install ID"
// @Param req body ResolveInstallDriftRequest true "Input"
// @Tags installs
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Success 200 {object} app.WorkflowResponse
// @Success 201 {object} app.WorkflowResponse
// @Router /v1/installs/{install_id}/resolve-drift [post]
func (s *service) ResolveInstallDrift(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req ResolveInstallDriftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(stderr.ErrUser{Err: err, Description: err.Error()})
		return
	}
	installID := ctx.Param("install_id")
	result, err := s.alignment.Resolve(ctx, org.ID, installID, (*alignment.ResolveRequest)(&req))
	if err != nil {
		ctx.Error(err)
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
		res := &result.Resource
		if deploy := result.Deploy; deploy != nil {
			s.emitInstallDeployAudit(ctx, installID, deploy.ID, res.ExpectedBuildID, res.ComponentID, res.ID, deploy.Type)
		}
		if response := result.ApprovalResponse; response != nil {
			approval := result.Approval
			s.logFlowAPIAction(ctx, "step.approval_response_submitted",
				zap.String("workflow_id", approval.InstallWorkflowStep.InstallWorkflowID), zap.String("step_id", approval.InstallWorkflowStepID),
				zap.String("approval_id", approval.ID), zap.String("approval_response_id", response.ID),
				zap.String("response_type", string(response.Type)), zap.String("resolve_drift_workflow_id", result.WorkflowID))
		}
		s.logFlowAPIAction(ctx, "workflow.resolve_drift_requested",
			zap.String("workflow_id", result.WorkflowID), zap.String("install_id", installID),
			zap.String("resource_id", res.ID), zap.String("resource_kind", string(res.Kind)), zap.String("category", string(res.Category)))
	}
	ctx.JSON(status, app.WorkflowResponse{WorkflowID: result.WorkflowID})
}
