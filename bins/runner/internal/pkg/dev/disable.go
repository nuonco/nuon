package dev

import (
	"os"
)

func (d *devver) Disabled() bool {
	switch d.watchRunnerType {
	case "org":
		return os.Getenv("DISABLE_ORG_RUNNER") == "true"
	case "install":
		return os.Getenv("DISABLE_INSTALL_RUNNER") == "true"
	default:
	}

	return false
}
