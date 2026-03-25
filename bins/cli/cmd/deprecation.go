package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/cli/styles"
)

func printDeprecatedCommandWarning(cmd *cobra.Command, deprecationMessage string) {
	msg := fmt.Sprintf("Command %q is deprecated, %s", cmd.CommandPath(), deprecationMessage)
	cmd.PrintErrln(styles.TextError.Bold(true).Render(msg))
}
