package oci

import (
	"context"
	"os"
)

func (w *oci) Cleanup(ctx context.Context) error {
	os.RemoveAll(w.tmpDir)
	return nil
}
