package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpRunActionInput struct {
	InstallID  string            `json:"install_id" jsonschema:"install ID to run the action on"`
	ActionID   string            `json:"action_id" jsonschema:"action workflow ID to run"`
	RunEnvVars map[string]string `json:"run_env_vars,omitempty" jsonschema:"optional run environment variables (keys become RUNENV_<key>)"`
	Role       string            `json:"role,omitempty" jsonschema:"optional custom/IAM role name for the run"`
}

type mcpRunActionResult struct {
	WorkflowID             string `json:"workflow_id"`
	InstallID              string `json:"install_id"`
	ActionWorkflowID       string `json:"action_workflow_id"`
	ActionWorkflowConfigID string `json:"action_workflow_config_id"`
}

func (s *service) mcpRunAction(ctx context.Context, _ *mcp.CallToolRequest, in mcpRunActionInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.InstallID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install: %w", err)
	}

	var configID string
	if install.AppConfigID != "" {
		awc, err := s.actionsHelpers.GetActionWorkflowConfig(ctx, in.ActionID, install.AppConfigID)
		if err == nil {
			configID = awc.ID
		}
	}
	if configID == "" {
		awc, err := s.getActionWorkflowLatestConfig(ctx, in.ActionID)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to resolve action config for %q: %w", in.ActionID, err)
		}
		configID = awc.ID
	}

	result, err := s.createInstallActionWorkflowRun(ctx, in.InstallID, CreateInstallActionWorkflowRunRequest{
		ActionWorkFlowConfigID: configID,
		RunEnvVars:             in.RunEnvVars,
		Role:                   in.Role,
	})
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpRunActionResult{
		WorkflowID:             result.WorkflowID,
		InstallID:              in.InstallID,
		ActionWorkflowID:       result.ActionWorkflowID,
		ActionWorkflowConfigID: result.ActionWorkflowConfigID,
	})
}
