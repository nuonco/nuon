package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpGetWorkflowInput struct {
	WorkflowID string `json:"workflow_id" jsonschema:"workflow ID"`
}

type mcpWorkflowSummary struct {
	ID                string                   `json:"id"`
	Type              string                   `json:"type"`
	Status            string                   `json:"status"`
	StatusDescription string                   `json:"status_description"`
	OwnerID           string                   `json:"owner_id"`
	OwnerName         string                   `json:"owner_name,omitempty"`
	CreatedAt         string                   `json:"created_at"`
	CompletedSteps    int                      `json:"completed_steps"`
	TotalSteps        int                      `json:"total_steps"`
	PendingApproval   *mcpPendingApprovalInfo  `json:"pending_approval,omitempty"`
	Steps             []mcpWorkflowStepSummary `json:"steps"`
}

type mcpPendingApprovalInfo struct {
	ApprovalID string `json:"approval_id"`
	StepName   string `json:"step_name"`
	Type       string `json:"type"`
}

type mcpWorkflowStepSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	StepTargetType string `json:"step_target_type,omitempty"`
	StepTargetID   string `json:"step_target_id,omitempty"`
	HasLogs        bool   `json:"has_logs,omitempty"`
	ExecutionTime  string `json:"execution_time,omitempty"`
}

func (s *service) mcpGetWorkflow(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetWorkflowInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)

	var workflow app.Workflow
	err := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Steps.Approval.Response").
		Preload("StepGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx asc")
		}).
		Where("id = ? AND org_id = ?", in.WorkflowID, orgID).
		First(&workflow).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find workflow %q: %w", in.WorkflowID, err)
	}

	summary := mcpWorkflowSummary{
		ID:                workflow.ID,
		Type:              string(workflow.Type),
		Status:            string(workflow.Status.Status),
		StatusDescription: workflow.Status.StatusHumanDescription,
		OwnerID:           workflow.OwnerID,
		OwnerName:         workflow.OwnerName,
		CreatedAt:         workflow.CreatedAt.String(),
	}

	for _, step := range workflow.Steps {
		summary.TotalSteps++
		stepStatus := string(step.Status.Status)
		if stepStatus == string(app.StatusSuccess) {
			summary.CompletedSteps++
		}

		stepSummary := mcpWorkflowStepSummary{
			ID:             step.ID,
			Name:           step.Name,
			Status:         stepStatus,
			StepTargetType: step.StepTargetType,
			StepTargetID:   step.StepTargetID,
			HasLogs:        stepHasLogs(step.StepTargetType),
		}
		if step.ExecutionTime > 0 {
			stepSummary.ExecutionTime = step.ExecutionTime.String()
		}
		summary.Steps = append(summary.Steps, stepSummary)

		if step.Approval != nil && step.Approval.Response == nil {
			summary.PendingApproval = &mcpPendingApprovalInfo{
				ApprovalID: step.Approval.ID,
				StepName:   step.Name,
				Type:       string(step.Approval.Type),
			}
		}
	}

	return apiPkg.MCPJSONResult(summary)
}

func stepHasLogs(targetType string) bool {
	switch targetType {
	case string(app.WorkflowStepTargetTypeInstallDeploys),
		string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns),
		string(app.WorkflowStepTargetTypeInstallSandboxRuns):
		return true
	default:
		return false
	}
}
