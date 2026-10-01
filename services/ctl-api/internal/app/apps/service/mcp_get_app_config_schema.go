package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/pkg/config/schema"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetAppConfigSchemaInput struct {
	Type string `json:"type,omitempty" jsonschema:"schema type such as helm, action, or runbook. Omit to list valid types."`
}

type mcpGetAppConfigSchemaResult struct {
	Type          string   `json:"type,omitempty"`
	Header        string   `json:"header,omitempty"`
	SuggestedPath string   `json:"suggested_path,omitempty"`
	Schema        any      `json:"schema,omitempty"`
	ValidTypes    []string `json:"valid_types,omitempty"`
	Message       string   `json:"message,omitempty"`
}

func (s *service) mcpGetAppConfigSchema(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetAppConfigSchemaInput) (*mcp.CallToolResult, any, error) {
	if _, err := require.Read(ctx); err != nil {
		return nil, nil, err
	}

	types := schema.GetSchemaTypes()
	sort.Strings(types)

	typ := strings.TrimSpace(in.Type)
	if typ == "" {
		return apiPkg.MCPJSONResult(mcpGetAppConfigSchemaResult{
			ValidTypes: types,
			Message:    "Pass type to load one file schema.",
		})
	}

	root, err := schema.LookupSchemaType(typ)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to load schema %q: %w", typ, err)
	}
	if root == nil {
		return apiPkg.MCPJSONResult(mcpGetAppConfigSchemaResult{
			ValidTypes: types,
			Message:    fmt.Sprintf("Unknown schema type %q.", typ),
		})
	}

	encoded, err := json.Marshal(root)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encode schema %q: %w", typ, err)
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, nil, fmt.Errorf("unable to encode schema %q: %w", typ, err)
	}

	canonical := strings.ReplaceAll(typ, "_", "-")
	return apiPkg.MCPJSONResult(mcpGetAppConfigSchemaResult{
		Type:          canonical,
		Header:        "#" + canonical,
		SuggestedPath: suggestedConfigPath(canonical),
		Schema:        decoded,
	})
}

func suggestedConfigPath(schemaType string) string {
	switch schemaType {
	case "helm", "terraform", "docker-build", "container-image", "job", "kubernetes-manifest", "pulumi":
		return "components/<name>.toml"
	case "action":
		return "actions/<name>.toml"
	case "runbook":
		return "runbooks/<name>.toml"
	case "install":
		return "installs/<name>.toml"
	case "branch":
		return "branch.toml"
	case "input", "input-group":
		return "inputs/<name>.toml"
	case "inputs":
		return "inputs.toml"
	case "permission":
		return "permissions/<name>.toml"
	case "permission-policy":
		return "permissions/policies/<name>.toml"
	case "permissions":
		return "permissions.toml"
	case "policy":
		return "policies/<name>.toml"
	case "policies":
		return "policies.toml"
	case "secret":
		return "secrets/<name>.toml"
	case "secrets":
		return "secrets.toml"
	case "kubernetes-context":
		return "kubernetes_contexts/<name>.toml"
	case "kubernetes-contexts":
		return "kubernetes_contexts.toml"
	case "break-glass":
		return "break_glass.toml"
	case "installs-config":
		return "installs.toml"
	default:
		return schemaType + ".toml"
	}
}
