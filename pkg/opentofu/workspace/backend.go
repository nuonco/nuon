package workspace

import (
	"context"
	"fmt"
	"path/filepath"
)

const (
	defaultBackendConfigFilename = "backend.json"
)

// LoadBackend loads the backend from the provided plugin
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

func (w *workspace) backendFilepath() string {
	return filepath.Join(w.root, defaultBackendConfigFilename)
}
