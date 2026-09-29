package workspace

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

const defaultTmpRootDir string = "/tmp"
const defaultDirPermissions fs.FileMode = 0o777

type GitSource struct {
	URL  string `json:"url" validate:"required"`
	Ref  string `json:"ref" validate:"required"`
	Path string `json:"path"`
}

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

type Workspace struct {
	v   *validator.Validate
	src *GitSource

	tmpRootDir string `validate:"required"`
	id         string `validate:"required"`

	l *zap.Logger `validate:"required"`

	cleanupBeforeInit bool
}

type Option func(*Workspace)

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

func WithGitSource(src *GitSource) Option {
	return func(w *Workspace) {
		w.src = src
	}
}

func WithID(id string) Option {
	return func(w *Workspace) {
		w.id = id
	}
}

func WithTmpRoot(root string) Option {
	return func(w *Workspace) {
		w.tmpRootDir = root
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(w *Workspace) {
		w.l = l
	}
}

func WithCleanup(cleanup bool) Option {
	return func(w *Workspace) {
		w.cleanupBeforeInit = cleanup
	}
}

func (w *Workspace) cleanupExistingDir() error {
	rootDir := w.rootDir()
	if _, err := os.Stat(rootDir); err == nil {
		w.l.Info("removing existing workspace directory",
			zap.String("path", rootDir))
		if err := os.RemoveAll(rootDir); err != nil {
			w.l.Error("unable to remove existing directory",
				zap.String("path", rootDir),
				zap.Error(err))
			return fmt.Errorf("unable to remove existing directory: %w", err)
		}
	}
	return nil
}

func (w *Workspace) Init(ctx context.Context) error {
	if w.cleanupBeforeInit {
		if err := w.cleanupExistingDir(); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(w.rootDir(), defaultDirPermissions); err != nil {
		w.l.Error("unable to initialize root dir", zap.Error(err))
		return fmt.Errorf("unable to initialize root dir: %w", err)
	}

	if w.src == nil || w.src.URL == "" {
		return nil
	}

	if err := w.clone(ctx); err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	return nil
}

func (w *Workspace) Cleanup(ctx context.Context) error {
	if err := os.RemoveAll(w.rootDir()); err != nil {
		return fmt.Errorf("unable to cleanup directory: %w", err)
	}
	return nil
}

func (w *Workspace) Root() string {
	return w.rootDir()
}

func (w *Workspace) SourceDir() string {
	if w.src != nil && w.src.Path != "" {
		return filepath.Join(w.rootDir(), w.src.Path)
	}
	return w.rootDir()
}

func (w *Workspace) AbsPath(path string) string {
	return filepath.Join(w.rootDir(), path)
}

func (w *Workspace) IsFile(path string) bool {
	fp := w.AbsPath(path)
	stat, err := os.Stat(fp)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

func (w *Workspace) IsDir(path string) bool {
	fp := w.AbsPath(path)
	stat, err := os.Stat(fp)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func (w *Workspace) RmDir(path string) error {
	if !w.IsDir(path) {
		return nil
	}
	return os.RemoveAll(w.AbsPath(path))
}

func (w *Workspace) IsExecutable(path string) bool {
	fp := w.AbsPath(path)
	stat, err := os.Stat(fp)
	if err != nil {
		return false
	}
	if stat.IsDir() {
		return false
	}
	return stat.Mode()&0o111 != 0
}

func (w *Workspace) rootDir() string {
	return filepath.Join(w.tmpRootDir, w.id)
}
