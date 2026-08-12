package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/errs"
)

type (
	cobraRunCommand          func(*cobra.Command, []string)
	cobraRunECommand         func(*cobra.Command, []string) error
	cobraRunECommandExitCode func(*cobra.Command, []string) (int, error)
)

// wrapCmd wraps all CLI commands, providing a central point to control error flow and handling.
func (c *cli) wrapCmd(f cobraRunECommand) cobraRunCommand {
	fn := c.sentryWrapCmd(f)
	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
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
	fn := c.sentryWrapCmd(wrapped)
	return func(cmd *cobra.Command, args []string) {
		_ = fn(cmd, args)
	}
}

func (c *cli) sentryWrapCmd(f cobraRunECommand) cobraRunECommand {
	return func(cmd *cobra.Command, args []string) error {
		eventname := strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], "_")
		err := f(cmd, args)
		if err != nil {
			tags := map[string]string{
				"cmd_args":  strings.Join(os.Args, " "),
				"cli_event": eventname,
			}
			errs.ReportToSentry(err, &errs.SentryErrOptions{
				Tags:   tags,
				UserID: c.cfg.UserID,
			})
		}

		return err
	}
}
