package api

import "github.com/modelcontextprotocol/go-sdk/mcp"

func boolPtr(b bool) *bool { return &b }

// MCPReadTool builds a read-only MCP tool with standard annotations.
func MCPReadTool(name, title, description string) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
			OpenWorldHint:  boolPtr(false),
		},
	}
}

// MCPWriteTool builds a mutating MCP tool. Description should still use the
// WRITE OPERATION: prefix so the CLI --allow-writes filter can hide it.
func MCPWriteTool(name, title, description string, destructive, idempotent bool) *mcp.Tool {
	return &mcp.Tool{
		Name:        name,
		Title:       title,
		Description: description,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: boolPtr(destructive),
			IdempotentHint:  idempotent,
			OpenWorldHint:   boolPtr(false),
		},
	}
}
