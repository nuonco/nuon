package remote

import (
	"context"
	"fmt"
	"os"
)

// Uninstall removes the binary Install wrote. The terraform remote installer
// removes its version-managed install dir via hc-install; tofudl hands back
// raw bytes, so we own the file we wrote and remove exactly that.
func (r *remote) Uninstall(ctx context.Context) error {
	if r.installedPath != "" {
		if err := os.Remove(r.installedPath); err != nil {
			return fmt.Errorf("unable to remove installed tofu at %s: %w", r.installedPath, err)
		}
		r.installedPath = ""
	}

	return nil
}
