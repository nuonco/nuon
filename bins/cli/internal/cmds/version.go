package cmds

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/spf13/cobra"
)

func (c *cli) registerVersion(ctx context.Context, rootCmd *cobra.Command) error {
	rootCmd.AddCommand(&cobra.Command{
		Use: "version",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Printf("%s\n", version.Version)
			return nil
		},
	})
	return nil
}
