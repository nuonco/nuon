package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
)

// Execute is essentially the init method of the CLI. It initializes all the components and composes them together.
func Execute() {
	// Construct a validator for the API client and the UI logger.
	v := validator.New()
	c := &cli{
		v:   v,
		ctx: context.Background(),
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
	err := rootCmd.ExecuteContext(c.ctx)
	if c.useSentry {
		// Sentry should be flushed just the once, just prior to program exit
		sentry.Flush(2 * time.Second)
	}
	if err != nil || c.err != nil {
		os.Exit(2)
	}
}
