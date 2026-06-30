package api

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPService interface {
	RegisterMCPTools(server *mcp.Server)
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
