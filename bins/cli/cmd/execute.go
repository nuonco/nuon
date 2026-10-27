package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/nuonco/nuon/bins/cli/internal/httpdebug"
	"github.com/nuonco/nuon/bins/cli/internal/telemetry"
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
	c.start = start

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
		fang.WithErrorHandler(fangErrorHandler),
	)

	telemetry.Send(c.cfg, err == nil, time.Since(start))

	if Debug {
		httpdebug.PrintSummary(os.Stderr, time.Since(start))
	}

	if err != nil {
		os.Exit(2)
	}
}

// fangErrorHandler matches fang's default box without the period fang appends
// to every error.
func fangErrorHandler(w io.Writer, styles fang.Styles, err error) {
	if f, ok := w.(interface{ Fd() uintptr }); ok && !term.IsTerminal(int(f.Fd())) {
		fmt.Fprintln(w, err.Error())
		return
	}
	fmt.Fprintln(w, styles.ErrorHeader.String())
	fmt.Fprintln(w, styles.ErrorText.Render(err.Error()))
	fmt.Fprintln(w)
	if fangUsageError(err) {
		fmt.Fprintln(w, lipgloss.JoinHorizontal(
			lipgloss.Left,
			styles.ErrorText.UnsetWidth().Render("Try"),
			styles.Program.Flag.Render(" --help "),
			styles.ErrorText.UnsetWidth().UnsetMargins().UnsetTransform().Render("for usage."),
		))
		fmt.Fprintln(w)
	}
}

func fangUsageError(err error) bool {
	message := err.Error()
	for _, prefix := range []string{
		"flag needs an argument:",
		"unknown flag:",
		"unknown shorthand flag:",
		"unknown command",
		"invalid argument",
	} {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}
