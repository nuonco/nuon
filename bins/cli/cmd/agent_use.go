package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/pkg/agentclient"
)

// applyAgentMode stores the detected agent on the config and turns prompts
// off. Attribution and state tracking read cfg.Agent instead of detecting again.
func (c *cli) applyAgentMode() {
	if c.cfg == nil {
		return
	}
	client, ok := agentclient.Detect()
	if !ok {
		return
	}
	c.cfg.Agent = client.Name
	c.cfg.Interactive = false
}

// recordAgentUse writes ~/.nuon.agents/<agent>.yaml and, the first time that
// file is created, prints the setup guide on stderr. Failures are ignored so
// a local state problem does not block the command.
func (c *cli) recordAgentUse() {
	if c.cfg == nil || c.cfg.Agent == "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	created, err := agentclient.Touch(agentclient.Dir(home), agentclient.Record{
		Agent:      c.cfg.Agent,
		CLIVersion: version.Version,
		AppID:      c.cfg.AppID,
	}, time.Now())
	if err != nil || !created {
		return
	}
	fmt.Fprint(os.Stderr, agentclient.FirstRunGuide(c.cfg.Agent))
}
