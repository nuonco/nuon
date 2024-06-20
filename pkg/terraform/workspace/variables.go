package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultVariablesFilename string = "variables.json"
)

// getEnvironment returns the current environment as a map
func (w *workspace) getEnvironment() map[string]string {
	envVars := make(map[string]string)
	for _, val := range os.Environ() {
		pieces := strings.SplitN(val, "=", 2)
		envVars[pieces[0]] = pieces[1]
	}

	return envVars
}

// mergeMaps merges b into a, in place.
func (w *workspace) mergeMaps(a map[string]string, bs ...map[string]string) map[string]string {
	for _, b := range bs {
		for k, v := range b {
			a[k] = v
		}
	}

	return a
}

// LoadVariables initializes a variable set
func (w *workspace) LoadVariables(ctx context.Context) error {
	w.envVars = w.getEnvironment()

	for _, vars := range w.Variables {
		if err := vars.Init(ctx); err != nil {
			return fmt.Errorf("unable to init variables: %w", err)
		}

		varEnvVars, err := vars.GetEnv(ctx)
		if err != nil {
			return fmt.Errorf("unable to get env variables: %w", err)
		}
		w.envVars = w.mergeMaps(w.envVars, varEnvVars)

		byts, err := vars.GetFile(ctx)
		if err != nil {
			return fmt.Errorf("unable to get file variables: %w", err)
		}
		if len(byts) < 1 {
			continue
		}

		if err := w.writeFile(defaultVariablesFilename, byts, defaultFilePermissions); err != nil {
			return fmt.Errorf("unable to write file: %w", err)
		}
	}

	return nil
}

func (w *workspace) varsFilepath() string {
	return filepath.Join(w.root, defaultVariablesFilename)
}
