package workflow

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
)

// safeWriteFile writes data to path without following a pre-existing symlink.
// An image-backed action's container shares the workspace and could leave a
// symlink at a predictable control-file path (a later step's script, the
// supervisor, the outputs file); os.WriteFile would follow it and write
// through to a host target. Steps run sequentially, so the prior container has
// exited before we write here: unlink any existing entry, then create the file
// exclusively so a planted symlink cannot be followed.
func safeWriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to clear existing path")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, perm)
	if err != nil {
		return errors.Wrap(err, "unable to create file")
	}
	defer f.Close()

	// OpenFile's mode is masked by the process umask, which would quietly drop
	// the group/other bits a container running as a non-root user needs.
	if err := f.Chmod(perm); err != nil {
		return errors.Wrap(err, "unable to set file mode")
	}

	if _, err := f.Write(data); err != nil {
		return errors.Wrap(err, "unable to write file")
	}
	return nil
}
