package profiles

import (
	"os"
	"strconv"
	"strings"
)

const (
	EnvEnableProfiling = "ENABLE_PROFILING"
	EnvProfilingPort   = "PROFILING_PORT"
)

func LoadOptionsFromEnv() ProfilerOptions {
	options := DefaultProfilerOptions()

	enableStr := os.Getenv(EnvEnableProfiling)
	if enableStr != "" {
		enableStr = strings.ToLower(enableStr)
		options.Enabled = enableStr == "true" || enableStr == "1" || enableStr == "yes"
	}

	portStr := os.Getenv(EnvProfilingPort)
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			options.Port = port
		}
	}

	return options
}
