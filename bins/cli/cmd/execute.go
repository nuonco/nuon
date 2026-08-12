package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/getsentry/sentry-go"
	"github.com/nuonco/nuon/bins/cli/internal/httpdebug"
	"github.com/spf13/cobra"
)

// Building the description calls the API, so only do it when root help will
// actually render it. Flags are parsed first so --config is respected.
func populateRootLongDescription(c *cli, rootCmd *cobra.Command, args []string) {
	rootCmd.InitDefaultHelpFlag()
	if err := rootCmd.ParseFlags(args); err != nil {
		return
	}

	positional := rootCmd.Flags().Args()
	switch {
	case len(positional) == 0:
	case len(positional) == 1 && positional[0] == "help":
	default:
		return
	}

	// rootCmd already loaded config from the default path.
	_ = c.initConfig()

	rootCmd.Long = c.getLongDescription()
}

// Execute is essentially the init method of the CLI. It initializes all the components and composes them together.
func Execute() {
	start := time.Now()

	c, err := NewCLI()
	if err != nil {
		os.Exit(2)
	}

	// Kill CLI immediately when user types Ctrl-C.
	// Including SIGTERM to ensure consistent behavior.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		os.Exit(1)
	}()

	rootCmd := c.rootCmd()
	populateRootLongDescription(c, rootCmd, os.Args[1:])

	err = fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
		fang.WithoutVersion(),
	)

	// Sentry should be flushed just the once, just prior to program exit
	if c.cfg != nil && !c.cfg.DisableTelemetry {
		sentry.Flush(c.cfg.CleanupTimeout)
	}

	if Debug {
		httpdebug.PrintSummary(os.Stderr, time.Since(start))
	}

	if err != nil {
		os.Exit(2)
	}
}
