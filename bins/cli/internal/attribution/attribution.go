// Package attribution stamps the CLI version and agent headers on a nuon-go client.
package attribution

import (
	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/pkg/agentclient"
)

// Client is the part of the nuon-go client that carries attribution headers.
type Client interface {
	SetClientVersion(version string)
	SetAgentClient(name string)
	SetUserAgent(userAgent string)
	SetCommand(command string)
}

// Apply sets the CLI version and, when cfg.Agent is set, X-Nuon-Agent,
// User-Agent, and X-Nuon-Command.
func Apply(api Client, cfg *config.Config) {
	api.SetClientVersion(version.Version)
	if cfg == nil || cfg.Agent == "" {
		return
	}
	api.SetAgentClient(cfg.Agent)
	api.SetUserAgent(agentclient.UserAgent(version.Version, cfg.Agent))
	if cfg.AgentCommand != "" {
		api.SetCommand(cfg.AgentCommand)
	}
}
