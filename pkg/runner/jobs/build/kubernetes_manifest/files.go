package kubernetes_manifest

import (
	"errors"
	"fmt"
	"os"
)

func (h *handler) writeManifest(content []byte) error {
	file, err := h.state.arch.FS().OpenFile(defaultManifestFilename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to write manifest file: %w", err)
	}
	_, writeErr := file.Write(content)
	return errors.Join(writeErr, file.Close())
}
