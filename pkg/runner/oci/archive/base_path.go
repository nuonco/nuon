package ociarchive

import "github.com/spf13/afero"

// BasePath is a display path, not an I/O capability; use FS for confined reads.
func (a *archive) BasePath() string {
	return a.basePath
}

// FS borrows the pinned output root; callers must not close it. Cleanup or a
// rejected extraction closes the root.
func (a *archive) FS() afero.Fs {
	if a.store == nil {
		return nil
	}
	return a.store.output
}
