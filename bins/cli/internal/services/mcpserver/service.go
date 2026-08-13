// Package mcpserver runs a stdio MCP server exposing Nuon operations as tools
// for LLM clients (Claude Code, Claude Desktop, etc). stdout carries the MCP
// protocol, so tools must never print; they call the SDK directly and return
// JSON text content.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type Service struct {
	api nuon.Client
	cfg *config.Config
}

func New(api nuon.Client, cfg *config.Config) *Service {
	return &Service{api: api, cfg: cfg}
}

func (s *Service) Run(ctx context.Context, allowWrites bool) error {
	return s.buildServer(allowWrites).Run(ctx, &mcp.StdioTransport{})
}

func (s *Service) buildServer(allowWrites bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nuon",
		Version: version.Version,
	}, nil)

	s.registerReadTools(server)
	if allowWrites {
		s.registerWriteTools(server)
	}

	return server
}
