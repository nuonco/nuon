package service

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

type mcpNextAction struct {
	Action    string         `json:"action"`
	Label     string         `json:"label"`
	Tool      string         `json:"tool,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// mcpWorkflowAllowsApproveAll matches the dashboard Approve all button.
// The workflow must still be prompting, still running, not a drift scan, and
// contain at least one approval step.
func mcpWorkflowAllowsApproveAll(workflow app.Workflow) bool {
	if workflow.ApprovalOption != app.InstallApprovalOptionPrompt {
		return false
	}
	if workflow.PlanOnly || workflow.Finished || workflow.Status.Status == app.StatusCancelled {
		return false
	}
	for _, step := range workflow.Steps {
		if step.ExecutionType == app.WorkflowStepExecutionTypeHidden || step.Retried {
			continue
		}
		if step.ExecutionType == app.WorkflowStepExecutionTypeApproval {
			return true
		}
	}
	return false
}

func mcpPendingApprovalActions(approvalID, workflowID string, allowApproveAll bool) []mcpNextAction {
	actions := []mcpNextAction{
		{
			Action: "review_plan_diff",
			Label:  "Review current plan diffs",
			Tool:   "get_approval_plan_diff",
			Arguments: map[string]any{
				"approval_id": approvalID,
				"offset":      0,
			},
		},
		{
			Action: "approve",
			Label:  "Approve this step",
			Tool:   "approve_step",
			Arguments: map[string]any{
				"approval_id": approvalID,
			},
		},
		{
			Action: "reject",
			Label:  "Reject this step",
			Tool:   "reject_step",
			Arguments: map[string]any{
				"approval_id": approvalID,
			},
		},
		{
			Action: "stop",
			Label:  "Stop without changing anything",
		},
	}
	return appendApproveAll(actions, workflowID, allowApproveAll)
}

func mcpPlanDiffActions(approvalID, workflowID string, allowApproveAll, hasMore bool, nextOffset int) []mcpNextAction {
	if hasMore {
		return []mcpNextAction{{
			Action: "continue_review",
			Label:  "Load the remaining plan diffs before replying. Do not show this action to the user.",
			Tool:   "get_approval_plan_diff",
			Arguments: map[string]any{
				"approval_id": approvalID,
				"offset":      nextOffset,
			},
		}}
	}
	actions := []mcpNextAction{
		{
			Action: "approve",
			Label:  "Approve this step",
			Tool:   "approve_step",
			Arguments: map[string]any{
				"approval_id": approvalID,
			},
		},
		{
			Action: "reject",
			Label:  "Reject this step",
			Tool:   "reject_step",
			Arguments: map[string]any{
				"approval_id": approvalID,
			},
		},
		{
			Action: "stop",
			Label:  "Stop without changing anything",
		},
	}
	return appendApproveAll(actions, workflowID, allowApproveAll)
}

func (s *service) mcpWorkflowsForApprovals(ctx context.Context, orgID string, approvals []app.WorkflowStepApproval) map[string]app.Workflow {
	ids := make([]string, 0, len(approvals))
	seen := map[string]struct{}{}
	for _, approval := range approvals {
		id := approval.InstallWorkflowStep.InstallWorkflowID
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return s.mcpWorkflowsByID(ctx, orgID, ids)
}

func (s *service) mcpWorkflowsByID(ctx context.Context, orgID string, ids []string) map[string]app.Workflow {
	out := map[string]app.Workflow{}
	if len(ids) == 0 {
		return out
	}

	var workflows []app.Workflow
	err := s.db.WithContext(ctx).
		Select("id", "approval_option", "plan_only", "finished_at", "status").
		Where("id IN ? AND org_id = ?", ids, orgID).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "install_workflow_id", "execution_type", "retried")
		}).
		Find(&workflows).Error
	if err != nil {
		s.l.Warn("failed to load workflows for approval actions", zap.Error(err))
		return out
	}
	for _, workflow := range workflows {
		out[workflow.ID] = workflow
	}
	return out
}

func mcpAwaitingRetryActions(step app.WorkflowStep) []mcpNextAction {
	if step.Retried || step.ExecutionType == app.WorkflowStepExecutionTypeHidden {
		return nil
	}
	if directive.Step(step.ResultDirective) != directive.StepAwaitRetry {
		return nil
	}

	workflowID := step.InstallWorkflowID
	if workflowID == "" {
		workflowID = step.WorkflowID
	}

	actions := []mcpNextAction{{
		Action: "show_step_logs",
		Label:  "Show step logs",
		Tool:   "get_workflow_step_logs",
		Arguments: map[string]any{
			"step_id": step.ID,
		},
	}}
	if workflowID == "" {
		return actions
	}
	if step.Retryable {
		actions = append(actions, mcpNextAction{
			Action: "retry",
			Label:  "Retry step",
			Tool:   "retry_step",
			Arguments: map[string]any{
				"workflow_id": workflowID,
				"step_id":     step.ID,
			},
		})
	}
	if step.Skippable {
		actions = append(actions, mcpNextAction{
			Action: "skip",
			Label:  "Skip step",
			Tool:   "skip_step",
			Arguments: map[string]any{
				"workflow_id": workflowID,
				"step_id":     step.ID,
			},
		})
	}
	actions = append(actions, mcpNextAction{
		Action: "cancel_workflow",
		Label:  "Cancel workflow",
		Tool:   "cancel_workflow",
		Arguments: map[string]any{
			"workflow_id": workflowID,
		},
	})
	return actions
}

func appendApproveAll(actions []mcpNextAction, workflowID string, allowApproveAll bool) []mcpNextAction {
	if !allowApproveAll || workflowID == "" {
		return actions
	}
	return append(actions, mcpNextAction{
		Action: "approve_all",
		Label:  "Approve all",
		Tool:   "approve_all",
		Arguments: map[string]any{
			"workflow_id": workflowID,
		},
	})
}
