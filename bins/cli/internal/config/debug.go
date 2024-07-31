package config

import "os"

func Debug() bool {
	return os.Getenv(defaultDebugEnvVar) == "true"
}
