package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/cockroachdb/errors/withstack"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/bubbles"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse"
	"github.com/powertoolsdev/mono/pkg/config/sync"
	"github.com/powertoolsdev/mono/pkg/errs"
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

	var cfgErr config.ErrConfig
	if errors.As(err, &cfgErr) {
		msg := fmt.Sprintf("%s %s", cfgErr.Description, cfgErr.Error())
		if cfgErr.Warning {
			fmt.Println(bubbles.WarningStyle.Render(msg))
			return cfgErr
		}

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
		fmt.Println(bubbles.ErrorStyle.Render(parseErr.Description))
		if parseErr.Err != nil {
			fmt.Println(bubbles.ErrorStyle.Render(parseErr.Err.Error()))
		}

		return parseErr
	}

	fmt.Println(bubbles.ErrorStyle.Render(err.Error()))
	return err
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
