package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// enforceMinCLI warns when this build is below the control plane's warn floor
// and stops the command when it is below the error floor. version, help, and
// docs print the message and continue. A failed config fetch does neither.
func (c *cli) enforceMinCLI(cmd *cobra.Command) error {
	if skipMinCLICheck(cmd) || c.apiClient == nil {
		return nil
	}

	cfg, err := c.apiClient.GetCLIConfig(cmd.Context())
	if err != nil || cfg == nil {
		return nil
	}

	warn, block := version.Assess(cfg)
	if warn != "" {
		ui.PrintWarning(warn)
	}
	if block == nil {
		return nil
	}
	block = &ui.CLIUserError{Msg: block.Error()}
	if advisoryMinCLI(cmd) {
		ui.PrintWarning(block.Error())
		return nil
	}
	if agentmode.Enabled() || Output == OutputJSON {
		_ = ui.PrintError(block)
		c.exit(1)
	}
	return block
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
		if current.Name() == "login" {
			return true
		}
	}
	return false
}

func advisoryMinCLI(cmd *cobra.Command) bool {
	for current := cmd; current != nil; current = current.Parent() {
		switch current.Name() {
		case "version", "help", "docs":
			return true
		}
	}
	return false
}
