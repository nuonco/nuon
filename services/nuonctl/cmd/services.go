package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/services"
	"github.com/spf13/cobra"
)

func (c *cli) registerServices(ctx context.Context, rootCmd *cobra.Command) error {
	cmds, err := services.New(c.v)
	if err != nil {
		return fmt.Errorf("unable to initialize service commands: %w", err)
	}

	var generalCmd = &cobra.Command{
		Use:     "service",
		Aliases: []string{"s"},
		Short:   "commands for developing and interacting with services",
	}
	rootCmd.AddCommand(generalCmd)

	var name string
	generalCmd.PersistentFlags().StringVar(&name, "name", "", "service name")
	generalCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "run a service locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.Run(ctx, name)
		},
	})

	generalCmd.AddCommand(&cobra.Command{
		Use:   "env",
		Short: "fetch a service's env and output as json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.Env(ctx, name)
		},
	})

	return nil
}
