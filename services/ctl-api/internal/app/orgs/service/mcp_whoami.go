package service

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpWhoamiInput struct{}

func (s *service) mcpWhoami(ctx context.Context, _ *mcp.CallToolRequest, _ mcpWhoamiInput) (*mcp.CallToolResult, any, error) {
	orgID := keys.OrgIDFromContext(ctx)
	accountID := keys.CreatedByIDFromContext(ctx)

	var acct app.Account
	if res := s.db.WithContext(ctx).First(&acct, "id = ?", accountID); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to fetch account: %w", res.Error)
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}

	return apiPkg.MCPJSONResult(map[string]any{
		"account": map[string]string{"id": acct.ID, "email": acct.Email},
		"org":     map[string]string{"id": org.ID, "name": org.Name},
	})
}
