package config

import (
	"os"

	"golang.org/x/term"
)

const NoTTYEnvVar string = "NUON_NO_TTY"

func NoTTY() bool {
	return os.Getenv(NoTTYEnvVar) == "true"
}

func IsInteractive() bool {
	if NoTTY() {
		return false
	}
	if _, ok := os.LookupEnv("CI"); ok {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}
