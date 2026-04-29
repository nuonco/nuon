package workspace

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"

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

// providerZipPlatformRE matches the `<os>_<arch>` suffix of a packed
// provider zip in a terraform filesystem mirror, e.g.
// `terraform-provider-azurerm_4.34.0_linux_amd64.zip` →
// captures "linux_amd64".
var providerZipPlatformRE = regexp.MustCompile(`_([a-z0-9]+_[a-z0-9]+)\.zip$`)

// DetectFilesystemMirror returns the path to pass to WithFilesystemMirror
// (relative to the workspace root), or "" if the unpacked archive at
// archBase does not contain a non-empty provider mirror tree at
// DefaultFilesystemMirrorDir, or if the mirror does not include the current
// runtime platform (runtime.GOOS_runtime.GOARCH).
//
// The install runner is intentionally feature-flag-unaware: whether or not
// providers are vendored is decided server-side at build time, and the
// install side only checks "did the artifact ship one we can use?".
// Callers should pass the result straight to WithFilesystemMirror — an
// empty string is a no-op and terraform init falls back to direct registry
// resolution.
//
// The platform check is a guard against cross-platform install runners
// (notably local-dev macOS): if the airgap mirror only ships linux_amd64 +
// linux_arm64 zips and the install runner is darwin_arm64, writing the
// `direct { exclude = ["*/*"] }` .terraformrc would deadlock terraform init.
// We'd rather fall back to direct registry resolution on the unsupported
// platform than fail loudly — production install runners are linux and
// will always find their platform in the mirror.
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

	platforms := MirrorPlatforms(archBase)
	want := runtime.GOOS + "_" + runtime.GOARCH
	for _, p := range platforms {
		if p == want {
			return DefaultFilesystemMirrorDir
		}
	}

	// Mirror present but no zips for the current platform. Skip it and let
	// terraform init resolve providers from the public registry. Callers
	// that want to surface this case in logs can call MirrorPlatforms
	// themselves.
	return ""
}

// MirrorPlatforms returns the sorted, de-duplicated set of `<os>_<arch>`
// platforms present in the filesystem mirror at
// archBase/DefaultFilesystemMirrorDir, derived from packed provider zip
// filenames. Returns nil if no mirror exists or it has no recognizable
// provider zips.
//
// Used both by DetectFilesystemMirror (to decide whether the current
// runtime platform is supported) and by callers that want to log the
// vendored platform set for diagnostics.
func MirrorPlatforms(archBase string) []string {
	mirrorDir := filepath.Join(archBase, DefaultFilesystemMirrorDir)

	seen := make(map[string]struct{})
	_ = filepath.WalkDir(mirrorDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		m := providerZipPlatformRE.FindStringSubmatch(d.Name())
		if len(m) == 2 {
			seen[m[1]] = struct{}{}
		}
		return nil
	})

	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
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
