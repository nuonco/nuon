package terraform

import (
	"os"
	"strings"
)

func getEnv() map[string]string {
	envVars := make(map[string]string, 0)
	for _, envVar := range os.Environ() {
		pair := strings.SplitN(envVar, "=", 2)
		envVars[pair[0]] = pair[1]
	}

	return envVars
}
