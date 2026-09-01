package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetInstallInputsInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpGetInstallInputsResult struct {
	InstallID   string            `json:"install_id"`
	InstallName string            `json:"install_name"`
	InputsID    string            `json:"inputs_id,omitempty"`
	Values      map[string]string `json:"values"`
	CreatedAt   string            `json:"created_at,omitempty"`
}

func (s *service) mcpGetInstallInputs(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetInstallInputsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get install %q: %w", in.Install, err)
	}

	result := mcpGetInstallInputsResult{
		InstallID:   install.ID,
		InstallName: install.Name,
		Values:      map[string]string{},
	}

	latest, err := s.getLatestInstallInputs(ctx, install.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apiPkg.MCPJSONResult(result)
		}
		return nil, nil, err
	}

	result.InputsID = latest.ID
	result.CreatedAt = latest.CreatedAt.String()
	result.Values = hstoreToStringMap(latest.Values)
	return apiPkg.MCPJSONResult(result)
}

func hstoreToStringMap(h pgtype.Hstore) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if v != nil {
			out[k] = *v
		}
	}
	return out
}
