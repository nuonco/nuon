package fxmodules

import (
	"go.uber.org/fx"

	mcpserver "github.com/nuonco/nuon/services/ctl-api/internal/app/mcp/server"
)

var MCPAPIModule = fx.Module("mcp-api",
	fx.Invoke(mcpserver.New),
)
