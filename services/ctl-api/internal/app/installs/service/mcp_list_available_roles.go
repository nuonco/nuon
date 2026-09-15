package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListAvailableRolesInput struct {
	Install       string `json:"install" jsonschema:"install name or ID"`
	OperationType string `json:"operation_type,omitempty" jsonschema:"operation: provision, reprovision, deprovision, deploy, teardown, or trigger. Use with principal_type to mark the default role"`
	PrincipalType string `json:"principal_type,omitempty" jsonschema:"principal: sandbox, component, or action"`
	PrincipalID   string `json:"principal_id,omitempty" jsonschema:"component ID or action workflow ID; required to mark default for component and action"`
}

type mcpAvailableRole struct {
	Name     string `json:"name"`
	ARN      string `json:"arn,omitempty"`
	RoleType string `json:"role_type"`
	Default  bool   `json:"default,omitempty"`
}

func (s *service) mcpListAvailableRoles(ctx context.Context, _ *mcp.CallToolRequest, in mcpListAvailableRolesInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}
	if err := validateRoleSelectionParams(in.PrincipalType, in.OperationType); err != nil {
		return nil, nil, err
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install %q: %w", in.Install, err)
	}

	roles, err := s.availableRolesForInstall(ctx, orgID, install, in.PrincipalType, in.OperationType, in.PrincipalID)
	if err != nil {
		return nil, nil, err
	}

	out := make([]mcpAvailableRole, 0, len(roles))
	for _, r := range roles {
		out = append(out, mcpAvailableRole{
			Name:     r.Name,
			ARN:      r.ARN,
			RoleType: r.RoleType,
			Default:  r.Default,
		})
	}
	return apiPkg.MCPJSONResult(out)
}
