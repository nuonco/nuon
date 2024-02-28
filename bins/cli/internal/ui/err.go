package ui

import (
	"os"

	"github.com/pterm/pterm"
)

func printDebugErr(err error) {
	if os.Getenv(debugEnvVar) == "" {
		return
	}

	pterm.Error.Println(err.Error())
}
