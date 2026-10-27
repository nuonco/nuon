package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// enforceMinCLI warns when this build is older than the control plane's
// min_cli_version. A failed config fetch does not warn.
func (c *cli) enforceMinCLI(cmd *cobra.Command) error {
	if skipMinCLICheck(cmd) || c.apiClient == nil {
		return nil
	}

	cfg, err := c.apiClient.GetCLIConfig(cmd.Context())
	if err != nil || cfg == nil {
		return nil
	}

	if msg := version.MinCLIWarning(cfg.MinCliVersion); msg != "" {
		ui.PrintWarning(msg)
	}
	return nil
}

// returnToFang leaves the error for fang on a terminal. Agent and JSON output
// print the envelope and exit before fang, which would otherwise print the box as well.
func (c *cli) returnToFang(err error) error {
	if agentmode.Enabled() || Output == OutputJSON {
		_ = ui.PrintError(err)
		c.exit(1)
	}
	return err
}

func (c *cli) runLogin(cmd *cobra.Command, _ []string) error {
	err := c.auth.Login(cmd.Context())
	if err == nil {
		return nil
	}
	return c.returnToFang(err)
}

func skipMinCLICheck(cmd *cobra.Command) bool {
	if cmd.Flags().Lookup("help") != nil && cmd.Flags().Changed("help") {
		return true
	}
	for current := cmd; current != nil; current = current.Parent() {
		switch current.Name() {
		case "version", "help", "docs", "login":
			return true
		}
	}
	return false
}
