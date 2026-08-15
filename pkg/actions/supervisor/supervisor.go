// Package supervisor provides the actions-supervisor: a small POSIX shell
// script Nuon mounts into an image-backed action's container and runs via the
// image's own /bin/sh. It installs the nuon_output helper and runs the
// rendered step script. Shipping it as a shell script (rather than a compiled
// binary) means it runs in any base image that has a shell — which
// image-backed actions require anyway, since the customer's inline_contents is
// itself a shell script — with no arch/libc/static-linking concerns.
package supervisor

import (
	_ "embed"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

//go:embed supervisor.sh
var Script []byte

const (
	// OutputFilepathEnvVar names the file the nuon_output helper appends to.
	OutputFilepathEnvVar = "NUON_ACTIONS_OUTPUT_FILEPATH"
	// RootEnvVar names the workspace root inside the container.
	RootEnvVar = "NUON_ACTIONS_ROOT"

	// Filename is the supervisor script name written into the workspace.
	Filename = ".nuon-actions-supervisor.sh"
)

// Write writes the embedded supervisor script into dir and returns its path.
// It refuses to follow a pre-existing symlink at the destination: the script
// lands in the shared workspace, where a prior image-backed action container
// could have planted one to redirect the write to a host file.
func Write(dir string) (string, error) {
	path := filepath.Join(dir, Filename)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, "unable to clear existing supervisor path")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0o755)
	if err != nil {
		return "", errors.Wrap(err, "unable to create actions supervisor script")
	}
	defer f.Close()

	if _, err := f.Write(Script); err != nil {
		return "", errors.Wrap(err, "unable to write actions supervisor script")
	}
	return path, nil
}
