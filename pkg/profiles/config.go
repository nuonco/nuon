package profiles

import (
	"os"
	"strconv"
	"strings"
)

const (
	// EnvEnableProfiling is the environment variable to control profiling
	EnvEnableProfiling = "ENABLE_PROFILING"
	// EnvProfilingPort is the environment variable to set profiling port
	EnvProfilingPort = "PROFILING_PORT"
)

// LoadOptionsFromEnv loads profiler options from environment variables
func LoadOptionsFromEnv() ProfilerOptions {
	options := DefaultProfilerOptions()

	// Check if profiling is enabled via environment variable
	enableStr := os.Getenv(EnvEnableProfiling)
	if enableStr != "" {
		enableStr = strings.ToLower(enableStr)
		options.Enabled = enableStr == "true" || enableStr == "1" || enableStr == "yes"
	}

	// Check for custom port
	portStr := os.Getenv(EnvProfilingPort)
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			options.Port = port
		}
	}

	return options
}
