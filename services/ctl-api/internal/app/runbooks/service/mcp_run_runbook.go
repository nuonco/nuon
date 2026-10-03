package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpRunRunbookInput struct {
	Install   string                          `json:"install" jsonschema:"install name or ID"`
	Runbook   string                          `json:"runbook" jsonschema:"runbook name or ID"`
	Inputs    map[string]*string              `json:"inputs,omitempty" jsonschema:"runbook input values"`
	Steps     []CreateRunbookRunStepSelection `json:"steps,omitempty" jsonschema:"omit to run every step. To skip steps, pass every step_id from get_runbook and set enabled false for the ones to skip. A step omitted from this list still runs. At least one step must be enabled"`
	Role      string                          `json:"role,omitempty" jsonschema:"optional IAM role name from list_available_roles; omit to use the default"`
	RequestID string                          `json:"request_id,omitempty" jsonschema:"optional idempotency key. The same id, runbook, inputs, and steps returns the original run. A different body, or an install that has moved to another app config, is rejected"`
}

type mcpRunRunbookResult struct {
	RunID            string         `json:"run_id"`
	WorkflowID       string         `json:"workflow_id"`
	WorkflowType     string         `json:"workflow_type"`
	WorkflowStatus   string         `json:"workflow_status"`
	InstallID        string         `json:"install_id"`
	InstallRunbookID string         `json:"install_runbook_id"`
	RunbookConfigID  string         `json:"runbook_config_id"`
	Status           string         `json:"status"`
	CreatedAt        string         `json:"created_at"`
	NextAction       *mcpNextAction `json:"next_action,omitempty"`
}

type mcpNextAction struct {
	Action    string         `json:"action"`
	Label     string         `json:"label"`
	Tool      string         `json:"tool,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
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

	install, err := s.resolveInstallRef(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}
	if err := s.installHelpers.ValidateInstallRole(ctx, install.ID, in.Role); err != nil {
		return nil, nil, err
	}

	triggered, err := s.createRunbookRun(ctx, orgID, accountID, install.ID, in.Runbook, CreateRunbookRunRequest{
		Inputs:    in.Inputs,
		Steps:     in.Steps,
		Role:      in.Role,
		RequestID: in.RequestID,
	})
	if err != nil {
		return nil, nil, err
	}

	workflowID := ""
	workflowStatus := string(app.StatusPending)
	if triggered.Workflow != nil {
		workflowID = triggered.Workflow.ID
		if triggered.Workflow.Status.Status != "" {
			workflowStatus = string(triggered.Workflow.Status.Status)
		}
	} else if triggered.Run.InstallWorkflowID != nil {
		workflowID = *triggered.Run.InstallWorkflowID
	}
	return apiPkg.MCPJSONResult(mcpRunRunbookResult{
		RunID:            triggered.Run.ID,
		WorkflowID:       workflowID,
		WorkflowType:     string(app.WorkflowTypeRunbookRun),
		WorkflowStatus:   workflowStatus,
		InstallID:        triggered.Run.InstallID,
		InstallRunbookID: triggered.Run.InstallRunbookID,
		RunbookConfigID:  triggered.Run.RunbookConfigID,
		Status:           string(triggered.Run.Status),
		CreatedAt:        apiPkg.MCPTime(triggered.Run.CreatedAt),
		NextAction:       mcpWatchNextAction(workflowID, workflowStatus),
	})
}

func mcpWatchNextAction(workflowID, lastKnownStatus string) *mcpNextAction {
	if workflowID == "" {
		return nil
	}
	args := map[string]any{"workflow_id": workflowID}
	if lastKnownStatus != "" {
		args["last_known_status"] = lastKnownStatus
	}
	return &mcpNextAction{
		Action:    "watch_workflow",
		Label:     "Watch workflow",
		Tool:      "watch_workflow",
		Arguments: args,
	}
}
