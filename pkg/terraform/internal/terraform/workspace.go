package terraform

import (
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	backendConfigFilename string = "backend.json"
	varsConfigFilename    string = "nuon.tfvars.json"
)

type workspaceWriter interface {
	GetWorkingDir() (string, error)
	GetWriter(string) (io.WriteCloser, error)
}

// TODO(jdt): document all functions
// TODO(jdt): document all inputs / required fields
// TODO(jdt): plumb logger throughout
type workspace struct {
	// ID is the opaque identifier for this run
	// historically has been the nuon install ID
	ID      string                 `validate:"required"`
	Module  *planv1.Object         `validate:"required,dive"`
	Backend *planv1.Object         `validate:"required,dive"`
	Vars    map[string]interface{} `validate:"required"`
	Version string                 `validate:"required,min=5"`

	// internal state
	validator       *validator.Validate
	tfExecPath      string
	workingDir      string
	workspaceWriter workspaceWriter
	tfExecutor      tfExecutor
	cleanupFns      []func() error
}

type workspaceOption func(*workspace) error

// NewWorkspace creates a new workspace
// Inspired by terraform cloud workspaces which give you an isolated place to run terraform operations
func NewWorkspace(v *validator.Validate, opts ...workspaceOption) (*workspace, error) {
	w := &workspace{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating workspace: validator is nil")
	}
	w.validator = v

	for _, opt := range opts {
		if err := opt(w); err != nil {
			return nil, err
		}
	}

	if err := w.validator.Struct(w); err != nil {
		return nil, err
	}

	return w, nil
}

func WithID(i string) workspaceOption {
	return func(w *workspace) error {
		w.ID = i
		return nil
	}
}

func WithBackendBucket(b *planv1.Object) workspaceOption {
	return func(w *workspace) error {
		w.Backend = b
		return nil
	}
}

func WithModuleBucket(b *planv1.Object) workspaceOption {
	return func(w *workspace) error {
		w.Module = b
		return nil
	}
}

func WithVars(m map[string]interface{}) workspaceOption {
	return func(w *workspace) error {
		w.Vars = m
		return nil
	}
}

func WithVersion(v string) workspaceOption {
	return func(w *workspace) error {
		w.Version = v
		return nil
	}
}
func (w *workspace) Cleanup() error {
	var errOut error
	for _, fn := range w.cleanupFns {
		if err := fn(); err != nil {
			errOut = multierror.Append(errOut, err)
		}
	}
	return errOut
}
