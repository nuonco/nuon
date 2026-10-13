package agentclient

import "fmt"

// FirstRunGuide is the one-time stderr notice printed the first time an agent runs the CLI.
func FirstRunGuide(agent string) string {
	return fmt.Sprintf(`Nuon agent setup (%s)

Docs        https://docs.nuon.co/guides/agents
MCP         nuon agents help
Dashboard   https://app.nuon.co

nuon auth login
      |
      v
nuon orgs select
      |
      +----> nuon apps select ----> nuon installs list
      |
      +----> nuon agents mcp
`, agent)
}
