package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (c *cli) versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:               "version",
		Short:             "Show the CLI version, the control plane API version, and min_cli_supported_version when set",
		PersistentPreRunE: c.versionPreRunE,
		Annotations:       skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := c.version
			return svc.Version(cmd.Context(), PrintJSON)
		}),
		GroupID: HelpGroup.ID,
	}

	return versionCmd
}

// versionPreRunE still prints the client version when the API URL is missing
// or the client cannot be constructed.
func (c *cli) versionPreRunE(cmd *cobra.Command, args []string) error {
	err := c.persistentPreRunE(cmd, args)
	if err == nil || c.version != nil {
		return err
	}
	var userErr *ui.CLIUserError
	if errors.As(err, &userErr) {
		return err
	}
	c.version = version.New(nil, c.cfg)
	return nil
}
