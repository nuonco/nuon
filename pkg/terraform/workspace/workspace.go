package workspace

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/terraform/archive"
	"github.com/powertoolsdev/mono/pkg/terraform/backend"
	"github.com/powertoolsdev/mono/pkg/terraform/binary"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
)

// Workspace exposes an interface for interacting with terraform and uses inputs to fetch source files, configure the
// backend, the binary and more.
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=interface_mock.go -source=interface.go -package=workspace
var _ Workspace = (*workspace)(nil)

type workspace struct {
	v *validator.Validate

	Archive   archive.Archive       `validate:"required"`
	Backend   backend.Backend       `validate:"required"`
	Variables []variables.Variables `validate:"required,min=1"`
	Binary    binary.Binary         `validate:"required"`
	Hooks     hooks.Hooks           `validate:"required"`

	DisableCleanup bool

	// internal vars for managing the workspace
	tmpDirRoot string
	root       string
	execPath   string
	envVars    map[string]string
}

type workspaceOption func(*workspace) error

func New(v *validator.Validate, opts ...workspaceOption) (*workspace, error) {
	w := &workspace{
		v:          v,
		tmpDirRoot: os.TempDir(),
		Variables:  make([]variables.Variables, 0),
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
