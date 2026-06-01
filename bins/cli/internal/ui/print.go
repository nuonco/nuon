package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors/withstack"
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

func PrintError(err error) error {
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
	fmt.Print(msg)
}

func PrintLn(msg string) {
	fmt.Println(bubbles.InfoStyle.Render(msg))
}

func PrintWarning(msg string) {
	fmt.Println(bubbles.WarningStyle.Render(msg))
}

func PrintSuccess(msg string) {
	fmt.Println(bubbles.SuccessStyle.Render(msg))
}

func PrintDebug(msg string) {
	if os.Getenv(debugEnvVar) != "true" {
		return
	}
	fmt.Println(bubbles.InfoStyle.Render("DEBUG: " + msg))
}
