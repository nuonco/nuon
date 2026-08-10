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

type mcpGetWorkflowStepInput struct {
	StepID string `json:"step_id" jsonschema:"workflow step ID"`
}

type mcpStepDetail struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	Status            string                   `json:"status"`
	StatusDescription string                   `json:"status_description,omitempty"`
	StepTargetType    string                   `json:"step_target_type,omitempty"`
	StepTargetID      string                   `json:"step_target_id,omitempty"`
	ExecutionType     string                   `json:"execution_type"`
	HasLogs           bool                     `json:"has_logs"`
	ExecutionTime     string                   `json:"execution_time,omitempty"`
	Retryable         bool                     `json:"retryable"`
	Skippable         bool                     `json:"skippable"`
	PendingApproval   *mcpPendingApprovalInfo  `json:"pending_approval,omitempty"`
	PolicyValidation  *mcpPolicyValidationInfo `json:"policy_validation,omitempty"`
	CreatedAt         string                   `json:"created_at"`
}

type mcpPolicyValidationInfo struct {
	Status string `json:"status"`
}

func (s *service) mcpGetWorkflowStep(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetWorkflowStepInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)

	var step app.WorkflowStep
	err := s.db.WithContext(ctx).
		Preload("Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Approval.Response").
		Preload("PolicyValidation").
		Where(app.WorkflowStep{OrgID: orgID}).
		Where("id = ?", in.StepID).
		First(&step).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find step %q: %w", in.StepID, err)
	}

	detail := mcpStepDetail{
		ID:                step.ID,
		Name:              step.Name,
		Status:            string(step.Status.Status),
		StatusDescription: step.Status.StatusHumanDescription,
		StepTargetType:    step.StepTargetType,
		StepTargetID:      step.StepTargetID,
		ExecutionType:     string(step.ExecutionType),
		HasLogs:           stepHasLogs(step.StepTargetType),
		Retryable:         step.Retryable,
		Skippable:         step.Skippable,
		CreatedAt:         step.CreatedAt.String(),
	}

	if step.ExecutionTime > 0 {
		detail.ExecutionTime = step.ExecutionTime.String()
	}

	if step.Approval != nil && step.Approval.Response == nil {
		detail.PendingApproval = &mcpPendingApprovalInfo{
			ApprovalID: step.Approval.ID,
			StepName:   step.Name,
			Type:       string(step.Approval.Type),
		}
	}

	if step.PolicyValidation != nil {
		detail.PolicyValidation = &mcpPolicyValidationInfo{
			Status: string(step.PolicyValidation.Status.Status),
		}
	}

	return apiPkg.MCPJSONResult(detail)
}
