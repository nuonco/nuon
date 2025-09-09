package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/getsentry/sentry-go"
)

// Execute is essentially the init method of the CLI. It initializes all the components and composes them together.
func Execute() {
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
	err = fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	)

	// Sentry should be flushed just the once, just prior to program exit
	if c.cfg != nil && !c.cfg.DisableTelemetry {
		sentry.Flush(c.cfg.CleanupTimeout)
		if c.analyticsClient != nil {
			c.analyticsClient.Close()
		}
	}

	if err != nil {
		os.Exit(2)
	}
}
