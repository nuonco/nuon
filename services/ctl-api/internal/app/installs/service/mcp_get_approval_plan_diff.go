package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type mcpGetApprovalPlanDiffInput struct {
	ApprovalID string `json:"approval_id" jsonschema:"approval ID from get_workflow or get_pending_approvals"`
	Limit      int    `json:"limit,omitempty" jsonschema:"maximum changes to return (default 20, max 100)"`
	Offset     int    `json:"offset,omitempty" jsonschema:"skip this many changes; use next_offset when has_more is true"`
	Action     string `json:"action,omitempty" jsonschema:"optional action filter: create, update, delete, replace, no-op, or read"`
	Search     string `json:"search,omitempty" jsonschema:"optional case-insensitive match against the resource address or kind"`
}

type mcpApprovalPlanDiffResult struct {
	ApprovalID  string              `json:"approval_id"`
	StepID      string              `json:"step_id,omitempty"`
	StepName    string              `json:"step_name,omitempty"`
	WorkflowID  string              `json:"workflow_id,omitempty"`
	Type        string              `json:"type"`
	Changes     []mcpPlanDiffChange `json:"changes"`
	HasMore     bool                `json:"has_more"`
	Limit       int                 `json:"limit"`
	Offset      int                 `json:"offset"`
	NextOffset  int                 `json:"next_offset,omitempty"`
	TooLarge    bool                `json:"too_large,omitempty"`
	NextActions []mcpNextAction     `json:"next_actions,omitempty"`
	mcpChangeCounts
}

func (s *service) mcpGetApprovalPlanDiff(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetApprovalPlanDiffInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.ApprovalID == "" {
		return nil, nil, fmt.Errorf("approval_id is required")
	}
	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	var approval app.WorkflowStepApproval
	err = s.db.WithContext(ctx).
		Preload("InstallWorkflowStep").
		Where("id = ? AND org_id = ?", in.ApprovalID, orgID).
		First(&approval).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find approval %q: %w", in.ApprovalID, err)
	}

	blobCtx := blobstore.WithBlobService(ctx, s.blobSvc)
	contents, _ := approval.GetContents(blobCtx, s.cfg.BlobReadEnabled)
	built, err := buildApprovalPlanDiff(approval.Type, contents, in.Action, in.Search)
	if err != nil {
		return nil, nil, err
	}

	page, hasMore := pagePlanDiffs(built.Changes, limit, offset)
	counts := mcpChangesFromApproval(&approval)
	if built.ChangesState != "" {
		counts.ChangesState = built.ChangesState
	}

	result := mcpApprovalPlanDiffResult{
		ApprovalID:      approval.ID,
		StepID:          approval.InstallWorkflowStepID,
		Type:            string(approval.Type),
		Changes:         page,
		HasMore:         hasMore,
		Limit:           limit,
		Offset:          offset,
		NextOffset:      apiPkg.MCPNextOffset(offset, limit, hasMore),
		TooLarge:        built.TooLarge,
		mcpChangeCounts: counts,
	}
	if approval.InstallWorkflowStep.ID != "" {
		result.StepName = approval.InstallWorkflowStep.Name
		result.WorkflowID = approval.InstallWorkflowStep.InstallWorkflowID
	}
	allowApproveAll := false
	if result.WorkflowID != "" {
		if workflow, ok := s.mcpWorkflowsByID(ctx, orgID, []string{result.WorkflowID})[result.WorkflowID]; ok {
			allowApproveAll = mcpWorkflowAllowsApproveAll(workflow)
		}
	}
	result.NextActions = mcpPlanDiffActions(approval.ID, result.WorkflowID, allowApproveAll, hasMore, result.NextOffset)
	return apiPkg.MCPJSONResult(result)
}
