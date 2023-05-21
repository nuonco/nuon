package workspace

import (
	"context"
	"fmt"
)

const (
	defaultVariablesFilename string = "variables.json"
)

func (w *workspace) LoadVariables(ctx context.Context) error {
	if err := w.Variables.Init(ctx); err != nil {
		return fmt.Errorf("unable to init variables: %w", err)
	}

	envVars, err := w.Variables.GetEnv(ctx)
	if err != nil {
		return fmt.Errorf("unable to get env variables: %w", err)
	}
	w.envVars = envVars

	byts, err := w.Variables.GetFile(ctx)
	if err != nil {
		return fmt.Errorf("unable to get file variables: %w", err)
	}

	if err := w.writeFile(defaultVariablesFilename, byts, defaultFilePermissions); err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}

	return nil
}
