package dev

import (
	"os"
)

func (d *devver) Disabled() bool {
	switch d.watchRunnerType {
	case "install":
		return os.Getenv("DISABLE_INSTALL_RUNNER") == "true"
	default:
	}

	return false
}
