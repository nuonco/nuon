package agentmode

import (
	"io"
	"os"
)

const EnvVar = "NUON_AGENT"

var enabled bool

func FromEnv() bool {
	v := os.Getenv(EnvVar)
	return v == "true" || v == "1"
}

func SetEnabled(v bool) { enabled = v }

func Enabled() bool { return enabled }

func HumanWriter() io.Writer {
	if enabled {
		return os.Stderr
	}
	return os.Stdout
}
