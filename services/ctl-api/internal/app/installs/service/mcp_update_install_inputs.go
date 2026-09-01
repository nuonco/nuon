package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpUpdateInstallInputsInput struct {
	Install          string            `json:"install" jsonschema:"install name or ID"`
	Inputs           map[string]string `json:"inputs" jsonschema:"input names to new values (partial update merged over current values)"`
	Role             string            `json:"role,omitempty" jsonschema:"optional custom/IAM role name for the update workflow"`
	DeployDependents *bool             `json:"deploy_dependents,omitempty" jsonschema:"deploy components that depend on the updated inputs (default true)"`
	InputsOnly       bool              `json:"inputs_only,omitempty" jsonschema:"save values without deploying components or reprovisioning the sandbox"`
	PlanOnly         bool              `json:"plan_only,omitempty" jsonschema:"if true, only plan the input-update workflow; do not apply"`
}

type mcpUpdateInstallInputsResult struct {
	InstallID   string `json:"install_id"`
	InstallName string `json:"install_name"`
	InputsID    string `json:"inputs_id"`
	WorkflowID  string `json:"workflow_id,omitempty"`
}

func (s *service) mcpUpdateInstallInputs(ctx context.Context, _ *mcp.CallToolRequest, in mcpUpdateInstallInputsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if len(in.Inputs) == 0 {
		return nil, nil, fmt.Errorf("inputs is required")
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install %q: %w", in.Install, err)
	}

	patch := make(map[string]*string, len(in.Inputs))
	for k, v := range in.Inputs {
		val := v
		patch[k] = &val
	}

	deployDependents := in.DeployDependents == nil || *in.DeployDependents

	inputs, err := s.applyInstallInputsUpdate(ctx, install, patch, in.Role, deployDependents, in.InputsOnly, in.PlanOnly, app.WorkflowTypeInputUpdate)
	if err != nil {
		return nil, nil, err
	}

	result := mcpUpdateInstallInputsResult{
		InstallID:   install.ID,
		InstallName: install.Name,
		InputsID:    inputs.ID,
	}
	if inputs.WorkflowID != nil {
		result.WorkflowID = *inputs.WorkflowID
	}
	return apiPkg.MCPJSONResult(result)
}
