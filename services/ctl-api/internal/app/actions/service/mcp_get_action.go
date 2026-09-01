package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetActionInput struct {
	InstallID string `json:"install_id" jsonschema:"install ID"`
	ActionID  string `json:"action_id" jsonschema:"action workflow ID (same as REST /installs/{id}/actions/{action_id})"`
}

type mcpActionDetail struct {
	ID                     string `json:"id"`
	InstallID              string `json:"install_id"`
	ActionWorkflowID       string `json:"action_workflow_id"`
	Name                   string `json:"name"`
	ActionWorkflowConfigID string `json:"action_workflow_config_id,omitempty"`
	CanTriggerManually     bool   `json:"can_trigger_manually"`
	CreatedAt              string `json:"created_at"`
}

func (s *service) mcpGetAction(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetActionInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.InstallID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install: %w", err)
	}

	iaw, err := s.getInstallActionWorkflow(ctx, in.InstallID, in.ActionID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install action: %w", err)
	}

	detail := mcpActionDetail{
		ID:               iaw.ID,
		InstallID:        iaw.InstallID,
		ActionWorkflowID: iaw.ActionWorkflowID,
		Name:             iaw.ActionWorkflow.Name,
		CreatedAt:        iaw.CreatedAt.String(),
	}

	if install.AppConfigID != "" {
		awc, err := s.actionsHelpers.GetActionWorkflowConfig(ctx, in.ActionID, install.AppConfigID)
		if err == nil {
			detail.ActionWorkflowConfigID = awc.ID
			detail.CanTriggerManually = awc.WorkflowConfigCanTriggerManually()
		}
	}

	if detail.ActionWorkflowConfigID == "" {
		awc, err := s.getActionWorkflowLatestConfig(ctx, in.ActionID)
		if err == nil {
			detail.ActionWorkflowConfigID = awc.ID
			detail.CanTriggerManually = awc.WorkflowConfigCanTriggerManually()
		}
	}

	return apiPkg.MCPJSONResult(detail)
}
