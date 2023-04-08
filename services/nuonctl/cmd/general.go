package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/commands/general"
	"github.com/spf13/cobra"
)

func (c *cli) registerGeneral(ctx context.Context, rootCmd *cobra.Command) error {
	cmds, err := general.New(c.v, general.WithTemporalRepo(c.temporal))
	if err != nil {
		return fmt.Errorf("unable to initialize general commands: %w", err)
	}

	var generalCmd = &cobra.Command{
		Use:     "general",
		Aliases: []string{"g"},
		Short:   "general commands for things like ids and more",
	}
	rootCmd.AddCommand(generalCmd)

	generalCmd.AddCommand(&cobra.Command{
		Use:   "new-id-pair",
		Short: "generate and print a new random unique id in long and short formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.NewIDPair()
		},
	})
	generalCmd.AddCommand(&cobra.Command{
		Use:   "new-short-id",
		Short: "generate and print a new short id",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.NewShortID()
		},
	})

	var inputID string
	generalCmd.PersistentFlags().StringVar(&inputID, "id", "", "input id")
	generalCmd.AddCommand(&cobra.Command{
		Use:   "to-short-id",
		Short: "convert an id into a short-id",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ToShortID(inputID)
		},
	})
	generalCmd.AddCommand(&cobra.Command{
		Use:   "to-long-id",
		Short: "convert an id into a long-id",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ToLongID(inputID)
		},
	})
	generalCmd.AddCommand(&cobra.Command{
		Use:   "kubecfg",
		Short: "get a kubecfg",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.GetKubecfg()
		},
	})

	generalCmd.AddCommand(&cobra.Command{
		Use:   "provision-canary",
		Short: "provision a canary workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.ProvisionCanary(ctx)
		},
	})

	generalCmd.AddCommand(&cobra.Command{
		Use:   "deprovision-canary",
		Short: "deprovision a canary workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmds.DeprovisionCanary(ctx)
		},
	})
	return nil
}
