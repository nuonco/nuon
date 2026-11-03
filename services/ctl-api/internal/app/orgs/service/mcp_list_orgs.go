package service

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type mcpListOrgsInput struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum orgs to return (default 20, max 100)"`
	Offset int    `json:"offset,omitempty" jsonschema:"skip this many orgs; use next_offset from a previous response when has_more is true"`
	Query  string `json:"query,omitempty" jsonschema:"optional case-insensitive match against org name or id"`
}

type mcpListOrgsResult struct {
	Orgs       []OrgListItem `json:"orgs"`
	HasMore    bool          `json:"has_more"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
	NextOffset int           `json:"next_offset,omitempty"`
}

func (s *service) mcpListOrgs(ctx context.Context, _ *mcp.CallToolRequest, in mcpListOrgsInput) (*mcp.CallToolResult, any, error) {
	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}
	orgs, err := s.getCustomerOrgs(ctx, in.Query, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	orgs, hasMore := apiPkg.MCPClipList(orgs, limit)
	return apiPkg.MCPJSONResult(mcpListOrgsResult{
		Orgs: orgs, HasMore: hasMore, Limit: limit, Offset: offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}
