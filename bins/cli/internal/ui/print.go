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

// PrintError renders an error in the active output mode: agent -> envelope,
// json -> JSON error object, table -> human-styled text.
func PrintError(err error) error {
	if agentEnabled() {
		return emitAgentError(err)
	}
	if jsonOutputEnabled() {
		return emitJSONError(err)
	}
	return printHumanError(err)
}

func printHumanError(err error) error {
	if os.Getenv(debugEnvVar) != "" {
		fmt.Println(bubbles.ErrorStyle.Render(fmt.Sprintf("DEBUG: %v", err)))
	}

	// Construct a stack trace if this error doesn't already have one
	if !errs.HasNuonStackTrace(err) {
		err = withstack.WithStackDepth(err, 1)
	}

	resolved := resolveError(err)
	if resolved.warning {
		fmt.Println(bubbles.WarningStyle.Render(resolved.msg))
	} else {
		fmt.Println(bubbles.ErrorStyle.Render(resolved.msg))
	}
	return resolved.err
}

// resolvedError is an error reduced to what a user should see. The JSON and human
// output paths share it so a config or parse failure reads the same in both.
type resolvedError struct {
	msg     string
	warning bool
	err     error
}

func resolveError(err error) resolvedError {
	cliUserErr := &CLIUserError{}
	if errors.As(err, &cliUserErr) {
		return resolvedError{msg: err.Error(), err: err}
	}

	if apiUserErr, ok := nuon.ToUserError(err); ok {
		return resolvedError{msg: apiUserErr.Description, err: err}
	}

	if nuon.IsServerError(err) {
		return resolvedError{msg: defaultServerErrorMessage, err: err}
	}

	// Handle any other API errors with a user-friendly message
	if apiErrMsg, ok := nuon.ToAPIError(err); ok {
		return resolvedError{msg: apiErrMsg, err: err}
	}

	var parseErr parse.ParseErr
	if errors.As(err, &parseErr) {
		return resolvedError{msg: parseErr.Error(), err: parseErr}
	}

	var cfgErr config.ErrConfig
	if errors.As(err, &cfgErr) {
		if cfgErr.Warning {
			// Warnings carry their (possibly multi-line) human message in Description; render it as-is.
			wmsg := cfgErr.Description
			if wmsg == "" {
				wmsg = cfgErr.Error()
			}
			return resolvedError{msg: wmsg, warning: true, err: cfgErr}
		}

		return resolvedError{msg: fmt.Sprintf("%s %s", cfgErr.Description, cfgErr.Error()), err: cfgErr}
	}

	var syncErr sync.SyncErr
	if errors.As(err, &syncErr) {
		return resolvedError{msg: syncErr.Error(), err: syncErr}
	}

	var syncAPIErr sync.SyncAPIErr
	if errors.As(err, &syncAPIErr) {
		return resolvedError{msg: syncAPIErr.Error(), err: syncAPIErr}
	}

	// Filter out ugly technical error messages that shouldn't be shown to users
	errMsg := err.Error()
	if containsTechnicalError(errMsg) {
		return resolvedError{msg: defaultUnknownErrorMessage, err: err}
	}

	return resolvedError{msg: errMsg, err: err}
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
