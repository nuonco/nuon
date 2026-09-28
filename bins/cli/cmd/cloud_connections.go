package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (c *cli) cloudConnectionsCmd() *cobra.Command {
	command := &cobra.Command{
		Use: "cloud-connections", Short: "Manage cloud connections",
		GroupID: AdditionalGroup.ID, PersistentPreRunE: c.persistentPreRunE,
	}
	outputs := outputsAnnotation(OutputTable, OutputJSON, OutputAgent)

	command.AddCommand(&cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List cloud connections", Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error { return c.cloudConnections.List(cmd.Context(), PrintJSON) }),
	})
	command.AddCommand(&cobra.Command{
		Use: "get <connection-id>", Short: "Get a cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Get(cmd.Context(), args[0], PrintJSON)
		}),
	})

	var name, platform, targetID, principal, defaultRegion, preset string
	create := &cobra.Command{
		Use: "create", Short: "Create a cloud connection", Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if targetID == "" {
				return ui.PrintError(fmt.Errorf("--target-id is required"))
			}
			if principal == "" {
				return ui.PrintError(fmt.Errorf("--principal is required"))
			}
			return c.cloudConnections.Create(cmd.Context(), name, platform, targetID, principal, defaultRegion, preset, PrintJSON)
		}),
	}
	create.Flags().StringVar(&name, "name", "", "Connection name")
	create.Flags().StringVar(&platform, "platform", "aws", "Cloud platform (aws only)")
	create.Flags().StringVar(&targetID, "target-id", "", "AWS account ID")
	create.Flags().StringVar(&principal, "principal", "", "AWS IAM role ARN")
	create.Flags().StringVar(&defaultRegion, "default-region", "", "Default AWS region")
	create.Flags().StringVar(&preset, "preset", "", "Access preset: stacks or custom (attach your own permissions policy)")
	_ = create.MarkFlagRequired("name")
	_ = create.MarkFlagRequired("preset")
	command.AddCommand(create)

	verify := &cobra.Command{
		Use: "verify <connection-id>", Short: "Verify a cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Verify(cmd.Context(), args[0], PrintJSON)
		}),
	}
	command.AddCommand(verify)

	command.AddCommand(&cobra.Command{
		Use: "delete <connection-id>", Short: "Delete an unused cloud connection", Args: cobra.ExactArgs(1), Annotations: outputs,
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return c.cloudConnections.Delete(cmd.Context(), args[0], PrintJSON)
		}),
	})
	return command
}
