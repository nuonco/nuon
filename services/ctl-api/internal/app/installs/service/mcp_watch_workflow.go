package service

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

const (
	watchDefaultTimeout = 30 * time.Second
	watchMaxTimeout     = 60 * time.Second
	watchPollInterval   = 3 * time.Second
)

type mcpWatchWorkflowInput struct {
	WorkflowID      string `json:"workflow_id" jsonschema:"workflow ID to watch"`
	LastKnownStatus string `json:"last_known_status,omitempty" jsonschema:"the status you already know about; returns when status differs from this"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty" jsonschema:"max seconds to wait for a change (default 30, max 60)"`
}

type mcpWatchWorkflowResult struct {
	Changed  bool               `json:"changed"`
	Workflow mcpWorkflowSummary `json:"workflow"`
}

func (s *service) mcpWatchWorkflow(ctx context.Context, _ *mcp.CallToolRequest, in mcpWatchWorkflowInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	timeout := watchDefaultTimeout
	if in.TimeoutSeconds > 0 {
		timeout = time.Duration(in.TimeoutSeconds) * time.Second
		if timeout > watchMaxTimeout {
			timeout = watchMaxTimeout
		}
	}

	deadline := time.Now().Add(timeout)

	for {
		summary, err := s.fetchWorkflowSummary(ctx, orgID, in.WorkflowID)
		if err != nil {
			return nil, nil, err
		}

		currentStatus := summary.Status
		if in.LastKnownStatus == "" || currentStatus != in.LastKnownStatus {
			return apiPkg.MCPJSONResult(mcpWatchWorkflowResult{
				Changed:  in.LastKnownStatus != "" && currentStatus != in.LastKnownStatus,
				Workflow: *summary,
			})
		}

		if time.Now().After(deadline) {
			return apiPkg.MCPJSONResult(mcpWatchWorkflowResult{
				Changed:  false,
				Workflow: *summary,
			})
		}

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(watchPollInterval):
		}
	}
}

func (s *service) fetchWorkflowSummary(ctx context.Context, orgID, workflowID string) (*mcpWorkflowSummary, error) {
	var workflow app.Workflow
	err := s.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Steps.Approval.Response").
		Where("id = ? AND org_id = ?", workflowID, orgID).
		First(&workflow).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find workflow %q: %w", workflowID, err)
	}

	summary := &mcpWorkflowSummary{
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

	return summary, nil
}
