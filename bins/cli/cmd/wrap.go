package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"

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
	fn := c.analyticsWrapCmd(f)
	fn = c.sentryWrapCmd(fn)

	return func(cmd *cobra.Command, args []string) {
		if err := fn(cmd, args); err != nil {
			os.Exit(1)
		}
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

		props := map[string]interface{}{
			"namespace": namespace,
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
		c.analyticsClient.Track(c.ctx, events.CliCommand, props)

		return err
	}
}

func (c *cli) sentryWrapCmd(f cobraRunECommand) cobraRunECommand {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil {
			errs.ReportToSentry(err)
		}

		return err
	}
}
