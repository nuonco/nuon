package agentmode

import (
	"io"
	"os"
)

// EnvVar enables agent mode without the --agent flag.
const EnvVar = "NUON_AGENT"

var enabled bool

// FromEnv reports whether NUON_AGENT requests agent mode.
func FromEnv() bool {
	v := os.Getenv(EnvVar)
	return v == "true" || v == "1"
}

func SetEnabled(v bool) { enabled = v }

func Enabled() bool { return enabled }

// HumanWriter is the destination for progress and human-facing output. In agent
// mode it is stderr, so stdout carries only the JSON result envelope.
func HumanWriter() io.Writer {
	if enabled {
		return os.Stderr
	}
	return os.Stdout
}
