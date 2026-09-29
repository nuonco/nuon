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

func (c *cli) wrapCmd(f cobraRunECommand) cobraRunCommand {
	return func(cmd *cobra.Command, args []string) {
		if err := f(cmd, args); err != nil {
			os.Exit(exitCodeForErr(err))
		}
	}
}

func exitCodeForErr(err error) int {
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) && ec.ExitCode() != 0 {
		return ec.ExitCode()
	}
	return 1
}

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
