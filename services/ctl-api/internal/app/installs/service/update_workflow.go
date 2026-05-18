package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
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
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
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

func (s *service) updateWorkflow(ctx *gin.Context, installWorkflowID string, req *UpdateWorkflowRequest) (*app.Workflow, error) {
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

	// Label the approval source on the workflow metadata when approve-all is set via the API
	if req.ApprovalOption != nil && *req.ApprovalOption == app.InstallApprovalOptionApproveAll {
		s.db.WithContext(ctx).Exec(
			`UPDATE install_workflows SET metadata = COALESCE(metadata, ''::hstore) || hstore('approval_type', 'approve-workflow') WHERE id = ?`,
			installWorkflowID,
		)
		s.db.WithContext(ctx).Exec(
			`UPDATE install_workflows SET status = jsonb_set(COALESCE(status::jsonb, '{}'::jsonb), '{metadata,approval_type}', '"approve-workflow"'::jsonb) WHERE id = ?`,
			installWorkflowID,
		)

		// Approve any steps currently stuck in awaiting-approval status.
		s.approveStuckSteps(ctx, installWorkflowID)
	}

	return &currentWorkflow, nil
}

// approveStuckSteps finds workflow steps in awaiting-approval status that have
// an approval without a response, and auto-approves them so the workflow can
// proceed immediately when approve-all is set.
func (s *service) approveStuckSteps(ctx *gin.Context, workflowID string) {
	var steps []app.WorkflowStep
	res := s.db.WithContext(ctx).
		Where("install_workflow_id = ?", workflowID).
		Where("status->>'status' IN ?", []string{string(app.AwaitingApproval), "awaiting-approval"}).
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Approval.Response").
		Find(&steps)
	if res.Error != nil {
		s.l.Warn("failed to query stuck steps for approve-all", zap.Error(res.Error))
		return
	}

	useQueues, _ := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)

	for _, step := range steps {
		if step.Approval == nil || step.Approval.Response != nil {
			continue
		}

		resp, err := s.createWorkflowStepApprovalResponse(ctx, step.Approval.ID, &CreateWorkflowStepApprovalResponseRequest{
			ResponseType: app.WorkflowStepApprovalResponseTypeApprove,
			Note:         "auto-approved (approve-all)",
		})
		if err != nil {
			s.l.Warn("failed to create approval response for stuck step",
				zap.String("step_id", step.ID),
				zap.Error(err))
			continue
		}

		if useQueues {
			if err := s.dispatchApprovalResponseSignal(ctx, workflowID, step.ID, step.Approval.ID, resp.ID, app.WorkflowStepApprovalResponseTypeApprove); err != nil {
				s.l.Warn("failed to dispatch approval response signal for stuck step",
					zap.String("step_id", step.ID),
					zap.Error(err))
			}
		}
	}
}
