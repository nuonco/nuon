package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpRunRunbookInput struct {
	Install string                          `json:"install" jsonschema:"install name or ID"`
	Runbook string                          `json:"runbook" jsonschema:"runbook name or ID"`
	Inputs  map[string]*string              `json:"inputs,omitempty" jsonschema:"runbook input values"`
	Steps   []CreateRunbookRunStepSelection `json:"steps,omitempty" jsonschema:"optional step selections; empty runs all steps"`
	Role    string                          `json:"role,omitempty" jsonschema:"optional custom or IAM role name"`
}

type mcpRunRunbookResult struct {
	RunID            string `json:"run_id"`
	WorkflowID       string `json:"workflow_id"`
	InstallID        string `json:"install_id"`
	InstallRunbookID string `json:"install_runbook_id"`
	RunbookConfigID  string `json:"runbook_config_id"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
}

func (s *service) mcpRunRunbook(ctx context.Context, _ *mcp.CallToolRequest, in mcpRunRunbookInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if in.Runbook == "" {
		return nil, nil, fmt.Errorf("runbook is required")
	}
	accountID := keys.CreatedByIDFromContext(ctx)
	if accountID == "" {
		return nil, nil, fmt.Errorf("authenticated account is required")
	}

	triggered, err := s.createRunbookRun(ctx, orgID, accountID, in.Install, in.Runbook, CreateRunbookRunRequest{
		Inputs: in.Inputs,
		Steps:  in.Steps,
		Role:   in.Role,
	})
	if err != nil {
		return nil, nil, err
	}

	workflowID := ""
	if triggered.Workflow != nil {
		workflowID = triggered.Workflow.ID
	} else if triggered.Run.InstallWorkflowID != nil {
		workflowID = *triggered.Run.InstallWorkflowID
	}
	return apiPkg.MCPJSONResult(mcpRunRunbookResult{
		RunID:            triggered.Run.ID,
		WorkflowID:       workflowID,
		InstallID:        triggered.Run.InstallID,
		InstallRunbookID: triggered.Run.InstallRunbookID,
		RunbookConfigID:  triggered.Run.RunbookConfigID,
		Status:           string(triggered.Run.Status),
		CreatedAt:        apiPkg.MCPTime(triggered.Run.CreatedAt),
	})
}
