package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
)

const (
	defaultWorkspaceFilePerms os.FileMode = 0600
)

// manager manages the writing to a terraform module directory.
type manager struct {
	// ID represents a unique id for the run and is used as part of the tmp dir
	ID string `validate:"required"`

	// internal state
	validator *validator.Validate
	pattern   string
	tmpDir    string
}

type managerOption func(*manager) error

// New initializes a new manager with the given options
func New(v *validator.Validate, opts ...managerOption) (*manager, error) {
	m := &manager{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating directory writer: validator is nil")
	}
	m.validator = v

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	if err := m.validator.Struct(m); err != nil {
		return nil, err
	}

	return m, nil
}

// WithID sets the ID for the workspace
// this is used in constructing the name of the temp directory
func WithID(id string) managerOption {
	return func(w *manager) error {
		w.ID = id
		w.pattern = fmt.Sprintf("nuon-module-%s", id)
		return nil
	}
}

// TODO(jdt): consider moving to internal/terraform/workspace ?
// Init initializes the workspace for writing
func (w *manager) Init(ctx context.Context) (func() error, error) {
	cleanup := w.cleanup

	err := w.createTmpDir()
	if err != nil {
		return cleanup, fmt.Errorf("unable to create tmp dir: %w", err)
	}

	return cleanup, nil
}

// GetWorkingDir returns the workspaces working directory if set or error
func (w *manager) GetWorkingDir() (string, error) {
	if w.tmpDir == "" {
		return "", fmt.Errorf("working directory unset")
	}
	return w.tmpDir, nil
}

// cleanup removes the temp directory for the workspace
func (w *manager) cleanup() error {
	return os.RemoveAll(w.tmpDir)
}

// createTmpDir: create a temporary directory for the workspace
func (w *manager) createTmpDir() error {
	dir, err := os.MkdirTemp(os.TempDir(), w.pattern)
	if err != nil {
		return err
	}

	w.tmpDir = dir
	return nil
}

// GetWriter will create an io.WriteCloser for a file in the workspace
// It will return a nil io.WriteCloser if the filename evaluates to the root of the workspace
func (w *manager) GetWriter(filename string) (io.WriteCloser, error) {
	fp, err := filepath.Abs(filepath.Join(w.tmpDir, filename))
	if err != nil {
		return nil, err
	}
	// NOTE(jdt): /tmp/something == /tmp/something
	if fp == w.tmpDir {
		return nil, nil
	}

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY, defaultWorkspaceFilePerms)
	if err != nil {
		return nil, err
	}

	return &workspaceWriteCloser{f: f}, nil
}

// workspaceWriteCloser is an abstraction over writing files to a workspace
type workspaceWriteCloser struct {
	f *os.File
}

func (w *workspaceWriteCloser) Write(byts []byte) (int, error) {
	return w.f.Write(byts)
}

func (w *workspaceWriteCloser) Close() error {
	return w.f.Close()
}
