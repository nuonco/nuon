package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpRunActionInput struct {
	Install                string            `json:"install" jsonschema:"install name or ID"`
	ActionID               string            `json:"action_id" jsonschema:"action workflow ID to run"`
	ActionWorkflowConfigID string            `json:"action_workflow_config_id,omitempty" jsonschema:"optional action workflow config ID; defaults to the install-pinned or latest config"`
	RunEnvVars             map[string]string `json:"run_env_vars,omitempty" jsonschema:"optional run environment variables (keys become RUNENV_<key>)"`
	Role                   string            `json:"role,omitempty" jsonschema:"optional IAM role name from list_available_roles; omit to use the default"`
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
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if in.ActionID == "" {
		return nil, nil, fmt.Errorf("action_id is required")
	}
	if err := validateRunEnvVarKeys(in.RunEnvVars); err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install: %w", err)
	}

	if err := s.installHelpers.ValidateInstallRole(ctx, install.ID, in.Role); err != nil {
		return nil, nil, err
	}

	configID, err := s.resolveRunActionConfigID(ctx, install.AppConfigID, in.ActionID, in.ActionWorkflowConfigID)
	if err != nil {
		return nil, nil, err
	}

	result, err := s.createInstallActionWorkflowRun(ctx, install.ID, CreateInstallActionWorkflowRunRequest{
		ActionWorkFlowConfigID: configID,
		RunEnvVars:             in.RunEnvVars,
		Role:                   in.Role,
	})
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(mcpRunActionResult{
		WorkflowID:             result.WorkflowID,
		InstallID:              install.ID,
		ActionWorkflowID:       result.ActionWorkflowID,
		ActionWorkflowConfigID: result.ActionWorkflowConfigID,
	})
}

// validateRunEnvVarKeys rejects keys the runner cannot export as environment
// variables once RUNENV_ is prepended.
func validateRunEnvVarKeys(envVars map[string]string) error {
	for k := range envVars {
		if k == "" {
			return fmt.Errorf("run_env_vars keys cannot be empty")
		}
		if strings.ContainsAny(k, "= \t\n") {
			return fmt.Errorf("run_env_vars key %q cannot contain '=' or whitespace", k)
		}
	}
	return nil
}

func (s *service) resolveRunActionConfigID(ctx context.Context, installAppConfigID, actionID, pinnedConfigID string) (string, error) {
	if pinnedConfigID != "" {
		awc, err := s.actionsHelpers.GetActionWorkflowConfigByID(ctx, pinnedConfigID)
		if err != nil {
			return "", fmt.Errorf("unable to get action workflow config %q: %w", pinnedConfigID, err)
		}
		if awc.ActionWorkflowID != actionID {
			return "", fmt.Errorf("action_workflow_config_id %q does not belong to action %q", pinnedConfigID, actionID)
		}
		return awc.ID, nil
	}

	if installAppConfigID != "" {
		awc, err := s.actionsHelpers.GetActionWorkflowConfig(ctx, actionID, installAppConfigID)
		if err == nil {
			return awc.ID, nil
		}
	}
	awc, err := s.getActionWorkflowLatestConfig(ctx, actionID)
	if err != nil {
		return "", fmt.Errorf("unable to resolve action config for %q: %w", actionID, err)
	}
	return awc.ID, nil
}
