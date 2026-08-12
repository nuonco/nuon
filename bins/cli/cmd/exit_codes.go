package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

type exitCode struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

var exitCodes = []exitCode{
	{Code: 0, Description: "Success"},
	{Code: 1, Description: "General error"},
	{Code: 2, Description: "CLI initialization or execution error"},
}

func (c *cli) exitCodesCmd() *cobra.Command {
	exitCodesCmd := &cobra.Command{
		Use:               "exit-codes",
		Short:             "Learn about exit codes",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			if PrintJSON {
				ui.PrintJSON(exitCodes)
				return nil
			}

			ui.Println("Exit codes:")
			for _, ec := range exitCodes {
				ui.Printf("  %d - %s\n", ec.Code, ec.Description)
			}
			return nil
		}),
		GroupID: HelpGroup.ID,
	}

	return exitCodesCmd
}
