package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/runbooks"
)

func (c *cli) runbooksCmd() *cobra.Command {
	var installID string

	runbooksCmd := &cobra.Command{
		Use:   "runbooks",
		Short: "Manage runbooks",
		Long: `Manage and view runbooks by install ID.

By default, launches an interactive TUI to view and run runbooks.`,
		Args:              cobra.NoArgs,
		Annotations:       tuiAnnotation(TUIAltScreen),
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           InstallGroup.ID,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := runbooks.New(c.apiClient, c.cfg)
			return svc.RunbooksTUI(cmd.Context(), installID, PrintJSON)
		}),
	}

	runbooksCmd.Flags().StringVarP(&installID, "install-id", "i", "", "The ID or name of the install")

	return runbooksCmd
}
