package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid"
	"github.com/spf13/cobra"
)

func (c *cli) registerIDs(_ context.Context, rootCmd *cobra.Command) {
	var idsCmd = &cobra.Command{
		Use:     "ids",
		Aliases: []string{"s"},
		Short:   "commands for working with UUIDs and shortids",
	}
	rootCmd.AddCommand(idsCmd)

	idsCmd.AddCommand(&cobra.Command{
		Use:   "new",
		Short: "generate and print a new random unique id in long and short formats",

		RunE: func(cmd *cobra.Command, args []string) error {
			short := shortid.New()
			long, err := shortid.ToUUID(short)
			if err != nil {
				return err
			}
			fmt.Printf("%s %s", long, short)
			return nil
		},
	})
}
