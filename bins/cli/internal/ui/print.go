package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors/withstack"
	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/errs"
)

const (
	defaultServerErrorMessage  string = "Oops, we have experienced a server error. Please try again in a few minutes."
	defaultUnknownErrorMessage string = "Oops, we have experienced an unexpected error. Please let us know about this."
	debugEnvVar                string = "NUON_DEBUG"
)

type CLIUserError struct {
	Msg string
}

func (u *CLIUserError) Error() string {
	return u.Msg
}

// ErrOrgNotSet / ErrAppNotSet are the canonical user-facing errors for commands
// that require a selected org/app. They classify as user_error under --output
// agent and render cleanly in json/table, replacing per-package helpers that
// printed to stdout and swallowed the error.
func ErrOrgNotSet() error {
	return &CLIUserError{Msg: "current org is not set, use `orgs select` to set one"}
}

func ErrAppNotSet() error {
	return &CLIUserError{Msg: "current app is not set, use `apps select` to set one"}
}

// printedError marks an error as already rendered so the command boundary
// (wrapCmd) doesn't render it a second time. Unwrap keeps errors.As/Is
// (ErrExitCode, CLIUserError, nuon API errors) working through the marker.
type printedError struct{ err error }

func (e *printedError) Error() string { return e.err.Error() }
func (e *printedError) Unwrap() error { return e.err }

func markPrinted(err error) error {
	var p *printedError
	if errors.As(err, &p) {
		return err
	}
	return &printedError{err: err}
}

// PrintError renders an error in the active output mode: agent -> envelope,
// json -> JSON error object, table -> human-styled text. It renders each
// error at most once — already-rendered errors pass through unchanged — so
// the command boundary can call it unconditionally and any code path may
// simply `return err` and rely on the boundary to render it.
func PrintError(err error) error {
	if err == nil {
		return nil
	}
	var p *printedError
	if errors.As(err, &p) {
		return err
	}
	if agentEnabled() {
		return &printedError{err: emitAgentError(err)}
	}
	if jsonOutputEnabled() {
		return &printedError{err: emitJSONError(err)}
	}
	return &printedError{err: printHumanError(err)}
}

// spinnerRendered finalizes an error a spinner Fail already rendered. In agent
// mode the spinner writes to stderr, so the error is returned unmarked and
// travels to the command boundary to become the stdout envelope.
func spinnerRendered(err error) error {
	if agentEnabled() {
		return err
	}
	return markPrinted(err)
}

func printHumanError(err error) error {
	if os.Getenv(debugEnvVar) != "" {
		fmt.Println(bubbles.ErrorStyle.Render(fmt.Sprintf("DEBUG: %v", err)))
	}

	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}

	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		fmt.Println(bubbles.ErrorStyle.Render(err.Error()))
		return err
	}

	apiUserErr, ok := nuon.ToUserError(err)
	if ok {
		fmt.Println(bubbles.ErrorStyle.Render(apiUserErr.Description))
		return err
	}

	if nuon.IsServerError(err) {
		fmt.Println(bubbles.ErrorStyle.Render(defaultServerErrorMessage))
		return err
	}

	// Handle any other API errors with a user-friendly message
	if apiErrMsg, ok := nuon.ToAPIError(err); ok {
		fmt.Println(bubbles.ErrorStyle.Render(apiErrMsg))
		return err
	}

	var cfgErr config.ErrConfig
	if errors.As(err, &cfgErr) {
		if cfgErr.Warning {
			// Warnings carry their (possibly multi-line) human message in Description; render it as-is.
			wmsg := cfgErr.Description
			if wmsg == "" {
				wmsg = cfgErr.Error()
			}
			fmt.Println(bubbles.WarningStyle.Render(wmsg))
			return cfgErr
		}

		msg := fmt.Sprintf("%s %s", cfgErr.Description, cfgErr.Error())
		fmt.Println(bubbles.ErrorStyle.Render(msg))
		return cfgErr
	}

	var syncErr sync.SyncErr
	if errors.As(err, &syncErr) {
		fmt.Println(bubbles.ErrorStyle.Render(syncErr.Error()))
		return syncErr
	}

	var syncAPIErr sync.SyncAPIErr
	if errors.As(err, &syncAPIErr) {
		fmt.Println(bubbles.ErrorStyle.Render(syncAPIErr.Error()))
		return syncAPIErr
	}

	var parseErr parse.ParseErr
	if errors.As(err, &parseErr) {
		fmt.Println(bubbles.ErrorStyle.Render(parseErr.Error()))
		if parseErr.Err != nil {
			fmt.Println(bubbles.ErrorStyle.Render(parseErr.Err.Error()))
		}

		return parseErr
	}

	// Filter out ugly technical error messages that shouldn't be shown to users
	errMsg := err.Error()
	if containsTechnicalError(errMsg) {
		fmt.Println(bubbles.ErrorStyle.Render(defaultUnknownErrorMessage))
		return err
	}

	fmt.Println(bubbles.ErrorStyle.Render(errMsg))
	return err
}

// containsTechnicalError checks if an error message contains technical details
// that shouldn't be shown to end users
func containsTechnicalError(msg string) bool {
	technicalPatterns := []string{
		"is not supported by the TextConsumer",
		"(*models.",
		"runtime.Consumer",
		"go-openapi",
	}
	for _, pattern := range technicalPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

func PrintRaw(msg string) {
	fmt.Fprint(agentmode.HumanWriter(), msg)
}

// Printf writes formatted human output to the mode-aware writer (stderr in agent
// mode, stdout otherwise), so it never pollutes the agent stdout envelope.
func Printf(format string, a ...any) {
	fmt.Fprintf(agentmode.HumanWriter(), format, a...)
}

// Println mirrors fmt.Println but routes to the mode-aware writer.
func Println(a ...any) {
	fmt.Fprintln(agentmode.HumanWriter(), a...)
}

// Print mirrors fmt.Print but routes to the mode-aware writer.
func Print(a ...any) {
	fmt.Fprint(agentmode.HumanWriter(), a...)
}

func PrintLn(msg string) {
	fmt.Fprintln(agentmode.HumanWriter(), bubbles.InfoStyle.Render(msg))
}

// PrintResult renders a command outcome in the active output mode: a JSON payload
// (raw JSON or agent envelope, via PrintJSON) when asJSON is set, otherwise a
// human status line. Used by mutating commands so json/agent callers get a
// machine-readable confirmation instead of an empty envelope.
func PrintResult(asJSON bool, humanMsg string, data any) {
	if asJSON {
		PrintJSON(data)
		return
	}
	PrintLn(humanMsg)
}

func PrintWarning(msg string) {
	fmt.Fprintln(agentmode.HumanWriter(), bubbles.WarningStyle.Render(msg))
}

func PrintSuccess(msg string) {
	fmt.Fprintln(agentmode.HumanWriter(), bubbles.SuccessStyle.Render(msg))
}

func PrintDebug(msg string) {
	if os.Getenv(debugEnvVar) != "true" {
		return
	}
	fmt.Fprintln(agentmode.HumanWriter(), bubbles.InfoStyle.Render("DEBUG: "+msg))
}
