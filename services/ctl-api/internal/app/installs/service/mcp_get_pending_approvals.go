package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetPendingApprovalsInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"maximum approvals to return (default 20, max 100)"`
	Offset int `json:"offset,omitempty" jsonschema:"skip this many approvals; use next_offset from a previous response when has_more is true"`
}

type mcpPendingApprovalItem struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	StepID      string          `json:"step_id,omitempty"`
	StepName    string          `json:"step_name,omitempty"`
	WorkflowID  string          `json:"workflow_id,omitempty"`
	CreatedAt   string          `json:"created_at"`
	NextActions []mcpNextAction `json:"next_actions,omitempty"`
	mcpChangeCounts
}

type mcpGetPendingApprovalsResult struct {
	Approvals  []mcpPendingApprovalItem `json:"approvals"`
	HasMore    bool                     `json:"has_more"`
	Limit      int                      `json:"limit"`
	Offset     int                      `json:"offset"`
	NextOffset int                      `json:"next_offset,omitempty"`
}

func (s *service) mcpGetPendingApprovals(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetPendingApprovalsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var approvals []app.WorkflowStepApproval
	err = s.db.WithContext(ctx).
		Select("id", "org_id", "type", "install_workflow_step_id", "created_at", "changes_state", "changes_create", "changes_update", "changes_delete", "changes_replace", "changes_noop").
		Omit("contents").
		Preload("InstallWorkflowStep", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "install_workflow_id")
		}).
		Where(app.WorkflowStepApproval{OrgID: orgID}).
		Where("id NOT IN (SELECT install_workflow_step_approval_id FROM install_workflow_step_approval_responses WHERE deleted_at = 0)").
		Order("created_at DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&approvals).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list pending approvals: %w", err)
	}

	approvals, hasMore := apiPkg.MCPClipList(approvals, limit)
	workflows := s.mcpWorkflowsForApprovals(ctx, orgID, approvals)

	out := make([]mcpPendingApprovalItem, 0, len(approvals))
	for _, a := range approvals {
		item := mcpPendingApprovalItem{
			ID:              a.ID,
			Type:            string(a.Type),
			CreatedAt:       apiPkg.MCPTime(a.CreatedAt),
			mcpChangeCounts: mcpChangesFromApproval(&a),
		}
		item.StepID = a.InstallWorkflowStepID
		if a.InstallWorkflowStep.ID != "" {
			item.StepName = a.InstallWorkflowStep.Name
			item.WorkflowID = a.InstallWorkflowStep.InstallWorkflowID
		}
		workflow := workflows[item.WorkflowID]
		item.NextActions = mcpPendingApprovalActions(item.ID, item.WorkflowID, mcpWorkflowAllowsApproveAll(workflow))
		out = append(out, item)
	}

	return apiPkg.MCPJSONResult(mcpGetPendingApprovalsResult{
		Approvals:  out,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
