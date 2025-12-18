package kubernetes_manifest

import (
	"fmt"
	"os"
)

func (h *handler) writeManifest(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("unable to write manifest file: %w", err)
	}
	return nil
}
