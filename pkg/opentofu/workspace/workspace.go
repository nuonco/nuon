// Package workspace implements an OpenTofu workspace for the `tofu` CLI.
//
// The concrete workspace type deliberately satisfies the terraform workspace
// interface, github.com/nuonco/nuon/pkg/terraform/workspace.Workspace,
// cross-package: pkg/terraform/run and pkg/terraform/teardown orchestrate
// plan/apply/destroy against that interface and contain no terraform-CLI
// calls, so they drive an OpenTofu workspace unchanged. interface.go holds
// the compile-time assertion.
//
// tfexec cannot drive tofu: its version parser requires a `Terraform vX`
// banner while OpenTofu prints `OpenTofu vX`, so tfexec cannot even
// initialize against the tofu binary. This package therefore executes tofu
// directly (see client.go) and imports terraform-exec and terraform-json for
// TYPES ONLY — tfexec.OutputMeta and tfjson.State/Plan/ValidateOutput are the
// return types of the shared interface's methods. It never constructs a
// tfexec.Terraform.
//
// The tool-agnostic pieces (archive, backend, hooks, variables) are reused
// from pkg/terraform unchanged; the run/teardown pipelines, the local binary
// installer (pkg/terraform/binary/local) and the binary.Binary interface are
// reused as-is too. Only the binary provisioning via tofudl and the tofu
// command execution live here.
package workspace

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/terraform/archive"
	"github.com/nuonco/nuon/pkg/terraform/backend"
	"github.com/nuonco/nuon/pkg/terraform/binary"
	"github.com/nuonco/nuon/pkg/terraform/hooks"
	"github.com/nuonco/nuon/pkg/terraform/variables"
)

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
	//   1. write a .tofurc into the workspace root that configures
	//      provider_installation { filesystem_mirror { path = "<abs path>" } direct { exclude = ["*/*"] } }
	//   2. set TF_CLI_CONFIG_FILE to point at that .tofurc
	//
	// The path may be relative or absolute. Relative paths are resolved
	// against the workspace root (which is created lazily). The
	// `direct { exclude = ["*/*"] }` block is the airgap guarantee:
	// tofu init will fail loudly if a provider is missing from the
	// mirror rather than silently fall back to the public registry.
	FilesystemMirrorPath string

	// internal vars for managing the workspace
	tmpDirRoot string
	root       string
	execPath   string
	envVars    map[string]string
	varsPaths  []string

	// client is the cached tofu CLI client. It is created lazily by
	// getClient once root, binary and env are known, and holds the cached
	// `tofu version` result so the CLI is version-probed at most once.
	// clientMu guards the lazy creation; GracefulShutdown runs on a monitor
	// goroutine concurrently with the job flow.
	client   *tofuClient
	clientMu sync.Mutex
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

// WithFilesystemMirror configures the workspace to consume providers from an
// OpenTofu filesystem mirror at the given path instead of downloading them
// from registry.opentofu.org. See the FilesystemMirrorPath field for details.
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
