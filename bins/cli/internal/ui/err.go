package ui

import (
	"fmt"
	"os"

	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
)

func printDebugErr(err error) {
	if os.Getenv(debugEnvVar) == "" {
		return
	}

	fmt.Println(bubbles.ErrorStyle.Render(err.Error()))
}
