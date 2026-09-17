package api

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

type MCPService interface {
	RegisterMCPTools(server *mcp.Server)
}

func AsNuonctlMCPService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(MCPService)),
		fx.ResultTags(`group:"nuonctl_mcp_services"`),
	)
}

func MCPJSONResult(v any) (*mcp.CallToolResult, any, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(j)}},
	}, nil, nil
}
