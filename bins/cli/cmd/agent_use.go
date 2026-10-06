package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/pkg/agentclient"
)

// recordAgentUse writes ~/.nuon.agents/<agent>.yaml and, the first time that
// file is created, prints the setup guide on stderr. Failures are ignored so
// a local state problem does not block the command.
func (c *cli) recordAgentUse() {
	client, ok := agentclient.Detect()
	if !ok || c.cfg == nil {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	created, err := agentclient.Touch(agentclient.Dir(home), agentclient.Record{
		Agent:      client.Name,
		CLIVersion: version.Version,
		AppID:      c.cfg.AppID,
	}, time.Now())
	if err != nil || !created {
		return
	}
	fmt.Fprint(os.Stderr, agentclient.FirstRunGuide(client.Name))
}
