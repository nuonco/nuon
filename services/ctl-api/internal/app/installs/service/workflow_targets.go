package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installlookup "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/lookup"
)

type workflowApprovalTarget struct {
	workflow *app.Workflow
	step     *app.WorkflowStep
	approval *app.WorkflowStepApproval
}

func (s *service) resolveWorkflow(ctx context.Context, orgID, workflowID string) (*app.Workflow, error) {
	return installlookup.Workflow(ctx, s.db, orgID, workflowID)
}

func (s *service) resolveWorkflowStep(ctx context.Context, orgID, workflowID, stepID string) (*app.Workflow, *app.WorkflowStep, error) {
	workflow, err := s.resolveWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return nil, nil, fmt.Errorf("workflow not found: %w", err)
	}

	var step app.WorkflowStep
	if err := s.db.WithContext(ctx).
		Where(&app.WorkflowStep{ID: stepID, OrgID: workflow.OrgID, InstallWorkflowID: workflow.ID}, "ID", "OrgID", "InstallWorkflowID").
		First(&step).Error; err != nil {
		return nil, nil, fmt.Errorf("workflow step not found: %w", err)
	}
	return workflow, &step, nil
}

func (s *service) resolveWorkflowApproval(ctx context.Context, orgID, approvalID string) (*workflowApprovalTarget, error) {
	var approval app.WorkflowStepApproval
	if err := s.db.WithContext(ctx).
		Where(&app.WorkflowStepApproval{ID: approvalID, OrgID: orgID}, "ID", "OrgID").
		Preload("InstallWorkflowStep", func(db *gorm.DB) *gorm.DB {
			return db.Where(&app.WorkflowStep{OrgID: orgID}, "OrgID")
		}).
		Preload("Response").
		First(&approval).Error; err != nil {
		return nil, fmt.Errorf("workflow step approval not found: %w", err)
	}

	step := &approval.InstallWorkflowStep
	if step.ID == "" {
		return nil, fmt.Errorf("workflow step not found: %w", gorm.ErrRecordNotFound)
	}
	workflow, err := s.resolveWorkflow(ctx, orgID, step.InstallWorkflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}
	return &workflowApprovalTarget{workflow: workflow, step: step, approval: &approval}, nil
}
