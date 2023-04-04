package cmd

import (
	"context"
	"fmt"
	"log"

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
		Run: func(_ *cobra.Command, args []string) {
			if err := cmds.Run(ctx, name, args); err != nil {
				log.Fatal(err)
			}
		},
	})

	generalCmd.AddCommand(&cobra.Command{
		Use:   "exec",
		Short: "execute a command with a service's config",
		RunE: func(_ *cobra.Command, args []string) error {
			return cmds.Exec(ctx, name, args)
		},
	})

	generalCmd.AddCommand(&cobra.Command{
		Use:   "env",
		Short: "fetch a service's stage env and output as json",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmds.Env(ctx, name); err != nil {
				log.Fatal(err)
			}
		},
	})

	return nil
}
