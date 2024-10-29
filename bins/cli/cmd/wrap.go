package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/powertoolsdev/mono/bins/cli/internal/version"
	"github.com/powertoolsdev/mono/pkg/analytics/events"
	"github.com/powertoolsdev/mono/pkg/errs"
)

type (
	cobraRunCommand  func(*cobra.Command, []string)
	cobraRunECommand func(*cobra.Command, []string) error
)

// wrapCmd wraps all CLI commands, providing a central point to control error flow and handling.
func (c *cli) wrapCmd(f cobraRunECommand) cobraRunCommand {
	return func(cmd *cobra.Command, args []string) {
		c.sentryWrapCmd(c.analyticsWrapCmd(f))
	}
}

func (c *cli) analyticsWrapCmd(f cobraRunECommand) cobraRunECommand {
	return func(cmd *cobra.Command, args []string) error {
		startTS := time.Now()
		err := f(cmd, args)

		namespace := "root"
		if cmd.Root() != nil {
			namespace = cmd.Root().Name()
		}

		flagsVisited := make([]string, 0)
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Changed { // Check if the flag was set by the user
				flagsVisited = append(flagsVisited, flag.Name)
			}
		})

		props := map[string]interface{}{
			"namespace": namespace,
			"cmd_args":  strings.Join(os.Args, " "),
			"command":   cmd.Name(),
			"latency":   time.Since(startTS).Seconds(),
			"status":    "ok",
			"version":   version.Version,
		}

		if err != nil {
			props["status"] = "error"
			props["error"] = err.Error()
		}

		c.analyticsClient.Identify(c.ctx)

		eventname := strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], "_")
		c.analyticsClient.Track(c.ctx, events.Event(eventname), props)

		return err
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
			errs.ReportToSentry(err, &tags)
		}

		return err
	}
}
