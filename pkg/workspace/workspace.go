package workspace

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/rootfs"
)

const defaultTmpRootDir string = "/tmp"
const defaultDirPermissions fs.FileMode = 0o700

// GitSource defines the git repository to clone.
type GitSource struct {
	URL  string `json:"url" validate:"required"`
	Ref  string `json:"ref" validate:"required"`
	Path string `json:"path"` // subdirectory within the repo
}

// GitSourceFromPlanTypes converts a plantypes.GitSource to a workspace.GitSource.
func GitSourceFromPlanTypes(src *plantypes.GitSource) *GitSource {
	if src == nil {
		return nil
	}
	return &GitSource{
		URL:  src.URL,
		Ref:  src.Ref,
		Path: src.Path,
	}
}

// Workspace provides a temporary directory with optional git clone support.
type Workspace struct {
	v   *validator.Validate
	src *GitSource

	tmpRootDir string `validate:"required"`
	id         string `validate:"required"`

	l *zap.Logger `validate:"required"`

	cleanupBeforeInit bool // Remove existing directory before Init()
	backing           afero.Fs
	fs                afero.Fs
	parent            *rootfs.FS
	owned             *rootfs.FS
	gitFS             *rootfs.Billy
	repo              *git.Repository
}

// Option configures a Workspace.
type Option func(*Workspace)

// New creates a new Workspace with the given options.
func New(v *validator.Validate, opts ...Option) (*Workspace, error) {
	l, _ := zap.NewProduction()
	w := &Workspace{
		l:          l,
		v:          v,
		tmpRootDir: defaultTmpRootDir,
	}

	for _, opt := range opts {
		opt(w)
	}
	if err := w.v.Struct(w); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return w, nil
}

// WithGitSource sets a git source for cloning.
func WithGitSource(src *GitSource) Option {
	return func(w *Workspace) {
		w.src = src
	}
}

// WithID sets the workspace ID (used as the temp directory name).
func WithID(id string) Option {
	return func(w *Workspace) {
		w.id = id
	}
}

// WithTmpRoot sets the root temp directory for the workspace.
func WithTmpRoot(root string) Option {
	return func(w *Workspace) {
		w.tmpRootDir = root
	}
}

// WithLogger sets the logger.
func WithLogger(l *zap.Logger) Option {
	return func(w *Workspace) {
		w.l = l
	}
}

// WithFS selects a pinned OS backend or a private in-memory namespace.
// The backing filesystem is borrowed and is never closed by the workspace.
func WithFS(fs afero.Fs) Option {
	return func(w *Workspace) { w.backing = fs }
}

// FS borrows the job filesystem until Close or Cleanup.
func (w *Workspace) FS() afero.Fs { return w.fs }

// Repository borrows the cloned repository until Close or Cleanup.
func (w *Workspace) Repository() (*git.Repository, error) {
	if w.repo == nil {
		return nil, fmt.Errorf("workspace has no open repository")
	}
	return w.repo, nil
}

// WithCleanup configures whether to remove the workspace directory before initialization.
// If true, Init() will delete any existing directory at the workspace path before creating it.
// This is useful when re-using workspace IDs or recovering from failed clones.
func WithCleanup(cleanup bool) Option {
	return func(w *Workspace) {
		w.cleanupBeforeInit = cleanup
	}
}

// cleanupExistingDir removes the workspace directory if it exists.
func (w *Workspace) cleanupExistingDir() error {
	if err := w.validatePaths(); err != nil {
		return err
	}
	return rootfs.RemoveFS(w.backing, w.tmpRootDir, w.id)
}

// Init creates the workspace directory and clones the git source if configured.
func (w *Workspace) Init(ctx context.Context) (retErr error) {
	if err := w.validatePaths(); err != nil {
		return err
	}
	if w.fs != nil {
		return fmt.Errorf("workspace already initialized")
	}
	parent, child, err := rootfs.PrepareFS(w.backing, w.tmpRootDir, w.id, w.cleanupBeforeInit)
	if err != nil {
		return fmt.Errorf("unable to initialize root dir: %w", err)
	}
	w.parent, w.fs = parent, child
	w.owned, _ = child.(*rootfs.FS)
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, w.Close())
		}
	}()

	if w.src == nil || w.src.URL == "" {
		return nil
	}

	if err := w.clone(ctx); err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	return nil
}

// Cleanup removes the workspace directory.
func (w *Workspace) Cleanup(ctx context.Context) error {
	var err error
	if w.parent != nil {
		err = w.parent.RemoveAll(w.id)
	} else if w.fs != nil {
		err = w.fs.RemoveAll(".")
	} else {
		err = w.cleanupExistingDir()
	}
	return errors.Join(err, w.Close())
}

// Close releases handles without removing the workspace directory.
func (w *Workspace) Close() error {
	var errs []error
	if w.gitFS != nil {
		errs = append(errs, w.gitFS.Close())
		w.gitFS = nil
	}
	w.repo = nil
	if w.owned != nil {
		errs = append(errs, w.owned.Close())
		w.owned = nil
	}
	if w.parent != nil {
		errs = append(errs, w.parent.Close())
		w.parent = nil
	}
	w.fs = nil
	return errors.Join(errs...)
}

func (w *Workspace) validatePaths() error {
	if !filepath.IsLocal(w.id) || w.id == "." || strings.ContainsAny(w.id, `/\`) {
		return fmt.Errorf("invalid workspace ID %q", w.id)
	}
	if w.src != nil {
		// A leading slash historically means repository-relative, not host-rooted.
		path := strings.TrimLeft(w.src.Path, "/")
		if path != "" && !filepath.IsLocal(path) {
			return fmt.Errorf("invalid repository path %q", w.src.Path)
		}
	}
	abs, err := filepath.Abs(w.tmpRootDir)
	if err == nil {
		w.tmpRootDir = abs
	}
	return err
}

// Root returns the workspace root directory path.
func (w *Workspace) Root() string {
	return w.rootDir()
}

// SourceDir returns the absolute path to the source subdirectory within the clone.
// If no git source path is configured, returns the root directory.
func (w *Workspace) SourceDir() string {
	if w.src != nil && w.src.Path != "" {
		return filepath.Join(w.rootDir(), w.src.Path)
	}
	return w.rootDir()
}

// AbsPath returns an absolute path within the workspace.
func (w *Workspace) AbsPath(path string) string {
	return filepath.Join(w.rootDir(), path)
}

// IsFile returns true if the path exists and is a file.
func (w *Workspace) IsFile(path string) bool {
	if w.fs == nil {
		return false
	}
	stat, err := w.fs.Stat(relativePath(path))
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

// IsDir returns true if the path exists and is a directory.
func (w *Workspace) IsDir(path string) bool {
	if w.fs == nil {
		return false
	}
	stat, err := w.fs.Stat(relativePath(path))
	if err != nil {
		return false
	}
	return stat.IsDir()
}

// RmDir removes a directory within the workspace.
func (w *Workspace) RmDir(path string) error {
	if !w.IsDir(path) {
		return nil
	}
	return w.fs.RemoveAll(relativePath(path))
}

// IsExecutable returns true if the path exists and is executable.
func (w *Workspace) IsExecutable(path string) bool {
	if w.fs == nil {
		return false
	}
	stat, err := w.fs.Stat(relativePath(path))
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}
	return stat.Mode()&0o111 != 0
}

func (w *Workspace) rootDir() string {
	if w.owned != nil {
		return w.owned.Root()
	}
	return filepath.Join(w.tmpRootDir, w.id)
}

func relativePath(name string) string {
	name = strings.TrimLeft(name, "/")
	if name == "" {
		return "."
	}
	return name
}
