package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

type (
	cobraRunCommand          func(*cobra.Command, []string)
	cobraRunECommand         func(*cobra.Command, []string) error
	cobraRunECommandExitCode func(*cobra.Command, []string) (int, error)
)

// wrapCmd wraps all CLI commands, providing a central point to control error flow and handling.
func (c *cli) wrapCmd(f cobraRunECommand) cobraRunCommand {
	return func(cmd *cobra.Command, args []string) {
		if err := f(cmd, args); err != nil {
			os.Exit(exitCodeForErr(err))
		}
	}
}

// exitCodeForErr resolves the process exit code for a command error. Errors
// carrying an ExitCode (e.g. ui.ErrExitCode) control their own code; anything
// else exits 1.
func exitCodeForErr(err error) int {
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) && ec.ExitCode() != 0 {
		return ec.ExitCode()
	}
	return 1
}

// wrapCmdWithExitCode wraps CLI commands that return custom exit codes.
// This is useful for commands like "watch" that need to signal different outcomes.
func (c *cli) wrapCmdWithExitCode(f cobraRunECommandExitCode) cobraRunCommand {
	wrapped := func(cmd *cobra.Command, args []string) error {
		exitCode, err := f(cmd, args)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return err
	}
	return func(cmd *cobra.Command, args []string) {
		_ = wrapped(cmd, args)
	}
}
