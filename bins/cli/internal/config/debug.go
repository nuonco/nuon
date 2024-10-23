package config

import "os"

const (
	DebugEnvVar string = "NUON_DEBUG"
)

func Debug() bool {
	return os.Getenv(DebugEnvVar) == "true"
}
