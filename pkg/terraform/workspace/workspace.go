package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/terraform/archive"
	"github.com/nuonco/nuon/pkg/terraform/backend"
	"github.com/nuonco/nuon/pkg/terraform/binary"
	"github.com/nuonco/nuon/pkg/terraform/hooks"
	"github.com/nuonco/nuon/pkg/terraform/variables"
)

// DefaultFilesystemMirrorDir is the conventional directory name (relative to
// a workspace root or an archive base path) that holds a terraform provider
// filesystem mirror. Build runners write the mirror into this directory and
// install runners pass the same path to WithFilesystemMirror.
//
// The leading dot avoids polluting `terraform fmt` output without colliding
// with terraform's own `.terraform/` working directory (which the runner
// ignores during archive packaging).
const DefaultFilesystemMirrorDir = ".terraform-providers"

// DetectFilesystemMirror returns the path to pass to WithFilesystemMirror
// (relative to the workspace root), or "" if the unpacked archive at
// archBase does not contain a non-empty provider mirror tree at
// DefaultFilesystemMirrorDir.
//
// The install runner is intentionally feature-flag-unaware: whether or not
// providers are vendored is decided server-side at build time, and the
// install side only checks "did the artifact ship one?". Callers should
// pass the result straight to WithFilesystemMirror — an empty string is a
// no-op and terraform init falls back to direct registry resolution.
//
// We return a workspace-relative path because the dirarchive copies the
// mirror into the workspace root; the workspace then resolves the absolute
// path against its own root, not the archive base.
func DetectFilesystemMirror(archBase string) string {
	mirrorDir := filepath.Join(archBase, DefaultFilesystemMirrorDir)

	entries, err := os.ReadDir(mirrorDir)
	if err != nil || len(entries) == 0 {
		// Common: dir doesn't exist (older artifact / flag off). Other
		// errors are also fine to ignore — terraform init falls back
		// to direct registry resolution and (with the .terraformrc
		// the workspace would have written) we'd never have written
		// one in this branch anyway.
		return ""
	}

	return DefaultFilesystemMirrorDir
}

// Workspace exposes an interface for interacting with terraform and uses inputs to fetch source files, configure the
// backend, the binary and more.
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=interface_mock.go -source=interface.go -package=workspace
var _ Workspace = (*workspace)(nil)

type workspace struct {
	v *validator.Validate

	Archive     archive.Archive       `validate:"required"`
	Backend     backend.Backend       `validate:"required"`
	Variables   []variables.Variables `validate:"required,min=1"`
	Binary      binary.Binary         `validate:"required"`
	Hooks       hooks.Hooks           `validate:"required"`
	PlanBytes   []byte
	PlanDisplay string

	DisableCleanup bool

	// FilesystemMirrorPath, when set, instructs the workspace to:
	//   1. write a .terraformrc into the workspace root that configures
	//      provider_installation { filesystem_mirror { path = "<abs path>" } direct { exclude = ["*/*"] } }
	//   2. set TF_CLI_CONFIG_FILE to point at that .terraformrc
	//
	// The path may be relative or absolute. Relative paths are resolved
	// against the workspace root (which is created lazily). The
	// `direct { exclude = ["*/*"] }` block is the airgap guarantee:
	// terraform init will fail loudly if a provider is missing from the
	// mirror rather than silently fall back to the public registry.
	FilesystemMirrorPath string

	// internal vars for managing the workspace
	tmpDirRoot string
	root       string
	execPath   string
	envVars    map[string]string
	varsPaths  []string
}

type workspaceOption func(*workspace) error

func New(v *validator.Validate, opts ...workspaceOption) (*workspace, error) {
	w := &workspace{
		v:          v,
		tmpDirRoot: os.TempDir(),
		Variables:  make([]variables.Variables, 0),
		varsPaths:  make([]string, 0),
	}

	for idx, opt := range opts {
		if err := opt(w); err != nil {
			return nil, fmt.Errorf("unable to set %d option: %w", idx, err)
		}
	}
	if err := w.v.Struct(w); err != nil {
		return nil, err
	}

	return w, nil
}

func WithArchive(arch archive.Archive) workspaceOption {
	return func(w *workspace) error {
		w.Archive = arch
		return nil
	}
}

func WithHooks(hooks hooks.Hooks) workspaceOption {
	return func(w *workspace) error {
		w.Hooks = hooks
		return nil
	}
}

func WithBackend(back backend.Backend) workspaceOption {
	return func(w *workspace) error {
		w.Backend = back
		return nil
	}
}

func WithVariables(vars variables.Variables) workspaceOption {
	return func(w *workspace) error {
		w.Variables = append(w.Variables, vars)
		return nil
	}
}

func WithBinary(bin binary.Binary) workspaceOption {
	return func(w *workspace) error {
		w.Binary = bin
		return nil
	}
}

func WithDisableCleanup(disable bool) workspaceOption {
	return func(w *workspace) error {
		w.DisableCleanup = disable
		return nil
	}
}

// WithFilesystemMirror configures the workspace to consume providers from a
// terraform filesystem mirror at the given path instead of downloading them
// from registry.terraform.io. See the FilesystemMirrorPath field for details.
func WithFilesystemMirror(path string) workspaceOption {
	return func(w *workspace) error {
		w.FilesystemMirrorPath = path
		return nil
	}
}

func WithPlanBytes(bytes []byte) workspaceOption {
	return func(w *workspace) error {
		w.PlanBytes = bytes
		return nil
	}
}
