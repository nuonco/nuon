package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/dev"
)

func (c *cli) devCmd() *cobra.Command {
	installID := ""
	yes := false
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Develop your app on Nuon",
		Long: `
A simple, guided experience for developing your app on Nuon.

Select an app and a dev install, then run this command to sync, build, and deploy to it.
`,
		PersistentPreRunE: c.persistentPreRunE,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			var dirName string
			if len(args) > 0 {
				dirName = args[0]
			} else {
				var err error
				dirName, err = os.Getwd()
				if err != nil {
					return errors.Wrap(err, "unable to get directory name")
				}
			}

			svc := dev.New(c.v, c.apiClient, c.cfg)
			return svc.Dev(cmd.Context(), dirName, installID, yes)
		}),
		GroupID: AdditionalGroup.ID,
	}
	devCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID of your dev install")
	devCmd.Flags().BoolVarP(&yes, "yes", "y", false, "If true, automatically approve all prompts")

	return devCmd
}
