package config

import (
	"os"

	"golang.org/x/term"
)

const NoTTYEnvVar string = "NUON_NO_TTY"

func NoTTY() bool {
	return os.Getenv(NoTTYEnvVar) == "true"
}

// IsInteractive returns true if the CLI should use interactive TUI mode.
// It checks (in priority order):
//  1. NUON_NO_TTY=true → non-interactive (explicit user override)
//  2. CI env var is set → non-interactive (standard CI convention)
//  3. stdout is not a terminal → non-interactive (pipe, redirect, cron)
//  4. Otherwise → interactive
func IsInteractive() bool {
	if NoTTY() {
		return false
	}
	if _, ok := os.LookupEnv("CI"); ok {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}
