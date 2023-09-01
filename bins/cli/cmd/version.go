package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/spf13/cobra"
)

func registerVersion(versionService *version.Service) cobra.Command {
	versionCmd := &cobra.Command{
		Use: "version",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return versionService.Version(cmd.Context())
		},
	}

	return *versionCmd
}
