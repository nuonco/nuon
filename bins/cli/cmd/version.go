package cmd

import (
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd(bindConfig config.BindCobraFunc, versionService *version.Service) *cobra.Command {
	versionCmd := &cobra.Command{
		Use: "version",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindConfig(cmd)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			versionService.Version(cmd.Context())
		},
	}

	return versionCmd
}
