package workspace

import (
	"context"
	"fmt"
)

const (
	defaultBackendConfigFilename = "backend.json"
)

func (w *workspace) LoadBackend(ctx context.Context) error {
	if err := w.Backend.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize backend: %w", err)
	}

	byts, err := w.Backend.ConfigFile(ctx)
	if err != nil {
		return fmt.Errorf("unable to get config file from backend: %w", err)
	}

	if err := w.writeFile(defaultBackendConfigFilename, byts, defaultFilePermissions); err != nil {
		return fmt.Errorf("unable to write config file: %w", err)
	}

	return nil
}
