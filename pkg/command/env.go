package command

import (
	"os"
	"strings"
)

func DefaultEnv() map[string]string {
	env := make(map[string]string)
	for _, envVar := range os.Environ() {
		envVarKV := strings.SplitN(envVar, "=", 2)
		k := envVarKV[0]
		v := envVarKV[1]

		env[k] = v
	}

	return env
}
