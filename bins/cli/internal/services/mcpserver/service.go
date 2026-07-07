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
	api         nuon.Client
	cfg         *config.Config
	allowWrites bool
}

func New(api nuon.Client, cfg *config.Config, allowWrites bool) *Service {
	return &Service{api: api, cfg: cfg, allowWrites: allowWrites}
}

func (s *Service) Run(ctx context.Context) error {
	return s.buildServer().Run(ctx, &mcp.StdioTransport{})
}

func (s *Service) buildServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "nuon",
		Version: version.Version,
	}, nil)

	s.registerReadTools(server)
	if s.allowWrites {
		s.registerWriteTools(server)
	}

	return server
}
