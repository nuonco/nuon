package cmd

import (
	"github.com/spf13/cobra"
)

func (c *cli) loginCmd() *cobra.Command {
	loginCmd := &cobra.Command{
		Deprecated:        "Use `nuon auth login` instead",
		Use:               "login",
		Short:             "Login to Nuon (deprecated)",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       skipAuthAnnotation(),
		RunE:              c.runLogin,
		GroupID:           AdditionalGroup.ID,
	}

	return loginCmd
}
