package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListComponentsInput struct {
	AppID string `json:"app_id,omitempty" jsonschema:"filter components by app ID"`
}

func (s *service) mcpListComponents(ctx context.Context, _ *mcp.CallToolRequest, in mcpListComponentsInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	var components []app.Component
	tx := s.db.WithContext(ctx).
		Where(app.Component{OrgID: orgID}).
		Order("name ASC")

	if in.AppID != "" {
		tx = tx.Where(app.Component{AppID: in.AppID})
	}

	if res := tx.Find(&components); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to list components: %w", res.Error)
	}

	return apiPkg.MCPJSONResult(components)
}
