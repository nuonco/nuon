// Package local implements a binary.Binary that points at an existing
// terraform CLI binary on disk. It is the install-runner counterpart of
// the build-time binary vendoring step: when the build runner shipped a
// terraform CLI inside the OCI artifact under
// pkg/terraform/workspace.DefaultBundledBinaryDir, the install runner
// instantiates this package's `local` binary instead of remotebinary so
// `terraform init`/`apply` can run fully airgapped.
//
// Lifecycle is intentionally tiny: no download, no copy, no cleanup.
// `Install` only validates the path is a regular executable file and
// returns it. The bundled binary lives inside the unpacked OCI archive,
// whose lifecycle is owned by the runner's archive layer — not by this
// binary implementation.
package local

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"

	"github.com/nuonco/nuon/pkg/terraform/binary"
)

type local struct {
	v *validator.Validate

	// Path is an absolute path to a terraform CLI binary that already
	// exists on disk and is executable. The local binary never modifies
	// or copies it.
	Path string `validate:"required"`
}

var _ binary.Binary = (*local)(nil)

type localOption func(*local) error

func New(v *validator.Validate, opts ...localOption) (*local, error) {
	l := &local{v: v}
	for _, opt := range opts {
		if err := opt(l); err != nil {
			return nil, err
		}
	}
	if err := l.v.Struct(l); err != nil {
		return nil, err
	}
	return l, nil
}

// WithPath sets the absolute path to the terraform CLI binary the
// workspace should execute. Caller is responsible for resolving and
// validating the path beforehand (see workspace.DetectBundledBinary).
func WithPath(p string) localOption {
	return func(l *local) error {
		l.Path = p
		return nil
	}
}

// Install satisfies binary.Binary. It verifies the configured path is a
// regular file, ensures it carries the executable bit, and returns it
// unchanged. The `dir` argument is ignored — local binaries live inside
// the OCI artifact, not under the workspace's bins/ dir.
//
// We chmod 0755 if the exec bit is missing rather than erroring out:
// OCI artifacts pulled via oras-go's file.Store do not preserve POSIX
// mode bits, so the bundled binary lands at the install runner with
// mode 0644 even though the build runner wrote 0755. Re-applying it
// here is the simplest place: we know we're about to exec the binary,
// and we own the chmod (no one else cares about the mode of a file
// inside an internal artifact dir).
func (l *local) Install(_ context.Context, _ hclog.Logger, _ string) (string, error) {
	info, err := os.Stat(l.Path)
	if err != nil {
		return "", fmt.Errorf("bundled terraform binary at %s: %w", l.Path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("bundled terraform binary path %s is a directory", l.Path)
	}
	if info.Mode()&0o111 == 0 {
		if err := os.Chmod(l.Path, 0o755); err != nil {
			return "", fmt.Errorf("unable to chmod bundled terraform binary at %s: %w", l.Path, err)
		}
	}
	return l.Path, nil
}

// Init is a no-op. The remote implementation uses Init for setup that
// runs once across an installer's lifetime; the local binary has nothing
// to do.
func (l *local) Init(_ context.Context) error {
	return nil
}

// Uninstall is a no-op. The bundled binary lives inside the unpacked OCI
// archive whose cleanup is owned elsewhere — removing it here would
// corrupt that archive for any later workspace built against the same
// base path.
func (l *local) Uninstall(_ context.Context) error {
	return nil
}
