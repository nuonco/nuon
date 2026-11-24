package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	currentWorkflow, err := s.resolveWorkflow(ctx, orgID, installWorkflowID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("workflow %s not found", installWorkflowID),
				Description: "workflow not found",
			}
		}
		return nil, fmt.Errorf("unable to get install workflow: %w", err)
	}

	res := s.db.WithContext(ctx).
		Model(currentWorkflow).
		Updates(req)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install workflow: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("install workflow not found: %w", gorm.ErrRecordNotFound)
	}

	// Label the approval source and release steps already waiting when approve-all is set.
	if req.ApprovalOption != nil && *req.ApprovalOption == app.InstallApprovalOptionApproveAll {
		s.markWorkflowApproveAll(ctx, currentWorkflow)
		s.logFlowAPIAction(ctx, "workflow.approve_all_requested",
			zap.String("workflow_id", installWorkflowID),
		)
	}

	return currentWorkflow, nil
}

// markWorkflowApproveAll records that this workflow was switched to approve-all
// and approves steps already waiting. Later plans in the workflow auto-approve.
func (s *service) markWorkflowApproveAll(ctx context.Context, workflow *app.Workflow) {
	s.db.WithContext(ctx).Exec(
		`UPDATE install_workflows SET metadata = COALESCE(metadata, ''::hstore) || hstore('approval_type', 'approve-workflow') WHERE id = ?`,
		workflow.ID,
	)
	s.db.WithContext(ctx).Exec(
		`UPDATE install_workflows SET status = jsonb_set(COALESCE(status::jsonb, '{}'::jsonb), '{metadata,approval_type}', '"approve-workflow"'::jsonb) WHERE id = ?`,
		workflow.ID,
	)
	s.approveStuckSteps(ctx, workflow)
}

// approveStuckSteps finds workflow steps in awaiting-approval status that have
// an approval without a response, and auto-approves them so the workflow can
// proceed immediately when approve-all is set.
func (s *service) approveStuckSteps(ctx context.Context, workflow *app.Workflow) {
	var steps []app.WorkflowStep
	res := s.db.WithContext(ctx).
		Where(&app.WorkflowStep{InstallWorkflowID: workflow.ID, OrgID: workflow.OrgID}, "InstallWorkflowID", "OrgID").
		Where("status->>'status' IN ?", []string{string(app.AwaitingApproval), "awaiting-approval"}).
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Where(&app.WorkflowStepApproval{OrgID: workflow.OrgID}, "OrgID").Omit("contents")
		}).
		Preload("Approval.Response").
		Find(&steps)
	if res.Error != nil {
		s.l.Warn("failed to query stuck steps for approve-all", zap.Error(res.Error))
		return
	}

	for _, step := range steps {
		if step.Approval == nil || step.Approval.Response != nil {
			continue
		}

		resp, err := s.createWorkflowStepApprovalResponse(ctx, step.Approval, &CreateWorkflowStepApprovalResponseRequest{
			ResponseType: app.WorkflowStepApprovalResponseTypeApprove,
			Note:         "auto-approved (approve-all)",
		})
		if err != nil {
			s.l.Warn("failed to create approval response for stuck step",
				zap.String("step_id", step.ID),
				zap.Error(err))
			continue
		}

		target := &workflowApprovalTarget{workflow: workflow, step: &step, approval: step.Approval}
		if err := s.dispatchApprovalResponseSignal(ctx, target, resp); err != nil {
			s.l.Warn("failed to dispatch approval response signal for stuck step",
				zap.String("step_id", step.ID),
				zap.Error(err))
		}
	}
}
