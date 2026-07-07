package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/workflowstepapprovalresponse"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
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
// @Failure					409	{object}	stderr.ErrResponse
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	workflowID := ctx.Param("workflow_id")
	stepID := ctx.Param("step_id")
	approvalID := ctx.Param("approval_id")

	_, err = s.getWorkflowStep(ctx, org.ID, workflowID, stepID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{
				Err:         err,
				Description: "workflow step not found",
			})
			return
		}
		ctx.Error(errors.Wrap(err, "unable to get workflow step"))
		return
	}

	approval, err := s.getWorkflowStepApproval(ctx, org.ID, approvalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{
				Err:         err,
				Description: "workflow step approval not found",
			})
			return
		}
		ctx.Error(errors.Wrap(err, "unable to get workflow step approval"))
		return
	}

	if approval.Response != nil {
		ctx.Error(stderr.ErrUser{
			Description: "workflow step approval already has a response",
			Err:         errors.New("workflow step approval already has a response"),
		})
		return
	}

	// create the response
	wfsaResponse, err := s.createWorkflowStepApprovalResponse(ctx, approval.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create workflow step approval response: %w", err))
		return
	}

	// Reactively unblock the step via the appropriate update.
	if req.ResponseType == app.WorkflowStepApprovalResponseTypeRetryPlan {
		// Retry is handled by the step-group (which clones and re-dispatches),
		// not by the approval flow inside the step signal.
		if _, err := s.flowsClient.RetryStep(ctx, &flowclient.RetryStepRequest{
			InstallWorkflowID: workflowID,
			StepID:            stepID,
		}); err != nil {
			s.l.Warn("failed to send retry-step update for approval retry", zap.Error(err))
		}
	} else {
		if err := s.dispatchApprovalResponseSignal(ctx, workflowID, stepID, approval.ID, wfsaResponse.ID, req.ResponseType); err != nil {
			s.l.Warn("failed to dispatch workflow-step-approval-response signal", zap.Error(err))
		}
	}

	response := CreateWorkflowStepApprovalResponseResponse{
		ID:   wfsaResponse.ID,
		Type: string(wfsaResponse.Type),
		Note: string(wfsaResponse.Note),
	}

	ctx.JSON(http.StatusCreated, response)
}

// dispatchApprovalResponseSignal forwards an approval response to the running workflow.
// For install-owned workflows, it enqueues a workflow-step-approval-response Nuon Signal
// for lifecycle webhooks and retries. For app-branch workflows, it calls flowsClient.ApprovePlan
// directly since the ApprovePlan activity is only registered on the installs worker.
func (s *service) dispatchApprovalResponseSignal(
	ctx *gin.Context,
	workflowID, stepID, approvalID, approvalResponseID string,
	responseType app.WorkflowStepResponseType,
) error {
	var wf app.Workflow
	if res := s.db.WithContext(ctx).Where(app.Workflow{ID: workflowID}).First(&wf); res.Error != nil {
		return fmt.Errorf("unable to load workflow %s: %w", workflowID, res.Error)
	}
	if wf.OwnerID == "" {
		return fmt.Errorf("workflow %s has no owner", workflowID)
	}

	switch wf.OwnerType {
	case "installs":
		return s.dispatchInstallApprovalSignal(ctx, wf.OwnerID, workflowID, stepID, approvalID, approvalResponseID, responseType)
	case "app_branches":
		return s.flowsClient.ApprovePlan(ctx, &flowclient.ApprovePlanRequest{
			InstallWorkflowID:  workflowID,
			StepID:             stepID,
			ApprovalResponseID: approvalResponseID,
			ResponseType:       responseType,
		})
	default:
		return fmt.Errorf("workflow %s has unsupported owner type %q", workflowID, wf.OwnerType)
	}
}

func (s *service) dispatchInstallApprovalSignal(
	ctx *gin.Context,
	installID, workflowID, stepID, approvalID, approvalResponseID string,
	responseType app.WorkflowStepResponseType,
) error {
	queueID, err := s.getInstallSignalsQueueID(ctx, installID)
	if err != nil {
		return fmt.Errorf("unable to resolve install-signals queue: %w", err)
	}

	sig := &workflowstepapprovalresponse.Signal{
		InstallID:          installID,
		InstallWorkflowID:  workflowID,
		WorkflowStepID:     stepID,
		ApprovalID:         approvalID,
		ApprovalResponseID: approvalResponseID,
		ResponseType:       responseType,
	}
	return s.enqueueInstallSignal(ctx, queueID, sig, approvalResponseID, "workflow_step_approval_responses")
}

func (s *service) createWorkflowStepApprovalResponse(ctx *gin.Context, approvalID string, req *CreateWorkflowStepApprovalResponseRequest) (*app.WorkflowStepApprovalResponse, error) {
	response := app.WorkflowStepApprovalResponse{
		InstallWorkflowStepApprovalID: approvalID,
		Type:                          req.ResponseType,
		Note:                          req.Note,
	}

	res := s.db.WithContext(ctx).Create(&response)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create workflow step approval response: %w", res.Error)
	}

	return &response, nil
}
